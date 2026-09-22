// Stockage objets S3-compatible (Cloudflare R2) pour les photos uploadées
// par l'app — voir cahier des charges technique, section "Stockage
// objets". Configuré uniquement par variables d'environnement, jamais en
// dur : S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY_ID, S3_SECRET_ACCESS_KEY,
// et optionnellement S3_REGION, S3_PUBLIC_BASE_URL.
package main

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ErrStorageNotConfigured signale que les variables d'environnement du
// stockage objets n'ont pas été fournies.
var ErrStorageNotConfigured = errors.New("stockage S3 non configuré : S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY_ID et S3_SECRET_ACCESS_KEY sont requis")

// ErrPostsStorageNotConfigured signale que le bucket public des visuels
// générés n'est pas configuré.
var ErrPostsStorageNotConfigured = errors.New("stockage des visuels générés non configuré : S3_POSTS_BUCKET et S3_POSTS_PUBLIC_URL sont requis")

// PhotoStorage encapsule l'accès aux buckets S3-compatibles utilisés par
// l'app : un bucket privé pour les photos sources des brouillons (bucket/
// publicBase, avec accès signé via presignedGetURL), et un bucket public
// séparé pour les visuels générés (postsBucket/postsPublicBase, exposés
// via le sous-domaine r2.dev — voir uploadPost), les deux accessibles via
// le même client/compte R2.
type PhotoStorage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	publicBase    string

	postsBucket     string
	postsPublicBase string
}

func newPhotoStorageFromEnv(ctx context.Context) (*PhotoStorage, error) {
	endpoint := os.Getenv("S3_ENDPOINT")
	bucket := os.Getenv("S3_BUCKET")
	accessKey := os.Getenv("S3_ACCESS_KEY_ID")
	secretKey := os.Getenv("S3_SECRET_ACCESS_KEY")
	if endpoint == "" || bucket == "" || accessKey == "" || secretKey == "" {
		return nil, ErrStorageNotConfigured
	}

	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "auto" // valeur attendue par Cloudflare R2
	}

	publicBase := os.Getenv("S3_PUBLIC_BASE_URL")
	if publicBase == "" {
		publicBase = strings.TrimSuffix(endpoint, "/") + "/" + bucket
	} else {
		publicBase = strings.TrimSuffix(publicBase, "/")
	}

	// Bucket séparé, public, pour les visuels générés — optionnel : si
	// absent, storeGeneratedImage échoue proprement (ErrPostsStorageNotConfigured)
	// plutôt que de faire planter le reste de l'API.
	postsBucket := os.Getenv("S3_POSTS_BUCKET")
	postsPublicBase := strings.TrimSuffix(os.Getenv("S3_POSTS_PUBLIC_URL"), "/")

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // requis par R2 et la plupart des S3-compatibles
	})
	presignClient := s3.NewPresignClient(client)

	return &PhotoStorage{
		client:        client,
		presignClient: presignClient,
		bucket:        bucket,
		publicBase:    publicBase,

		postsBucket:     postsBucket,
		postsPublicBase: postsPublicBase,
	}, nil
}

// upload envoie le contenu vers le bucket et renvoie l'URL publique de
// l'objet stocké. Cette URL n'est pas forcément accessible telle quelle si
// le bucket est privé — voir presignedGetURL.
func (s *PhotoStorage) upload(ctx context.Context, key, contentType string, body io.Reader) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}
	return s.publicBase + "/" + key, nil
}

// uploadPost envoie le contenu vers le bucket public dédié aux visuels
// générés (S3_POSTS_BUCKET) et renvoie son URL publique définitive,
// construite à partir de S3_POSTS_PUBLIC_URL. Contrairement à upload, pas
// de signature ni de TTL : cette URL est ce que le commerçant partage sur
// WhatsApp/Facebook, elle doit rester valide indéfiniment.
func (s *PhotoStorage) uploadPost(ctx context.Context, key, contentType string, body io.Reader) (string, error) {
	if s.postsBucket == "" || s.postsPublicBase == "" {
		return "", ErrPostsStorageNotConfigured
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.postsBucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}
	return s.postsPublicBase + "/" + key, nil
}

// keyFromPublicURL retrouve la clé objet à partir d'une URL renvoyée par
// upload (publicBase + "/" + key), pour pouvoir la re-signer plus tard.
// Renvoie false si l'URL ne vient pas de ce stockage (bucket/base changés
// entre l'upload et l'appel, par ex.).
func (s *PhotoStorage) keyFromPublicURL(url string) (string, bool) {
	prefix := s.publicBase + "/"
	if !strings.HasPrefix(url, prefix) {
		return "", false
	}
	return strings.TrimPrefix(url, prefix), true
}

// presignedGetURL renvoie une URL de téléchargement signée et à durée de
// vie limitée pour l'objet à la clé donnée. Le bucket R2 n'étant pas
// public, c'est cette URL (et non l'URL "publique" brute) qu'il faut
// transmettre à un tiers, comme le moteur de génération, pour qu'il
// puisse télécharger la photo.
func (s *PhotoStorage) presignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
