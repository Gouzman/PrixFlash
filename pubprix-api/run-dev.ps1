# Charge .env.local dans l'environnement puis lance le serveur.
# .env.local n'est jamais lu automatiquement par le programme Go — sans
# ce script (ou un `export`/`source` manuel équivalent), photoStorage
# reste nil et /drafts/{id}/photos renvoie 500 "stockage S3 non configuré".

$envFile = Join-Path $PSScriptRoot ".env.local"

if (-not (Test-Path $envFile)) {
    Write-Error "Fichier introuvable: $envFile (voir .env.local pour le modele attendu)"
    exit 1
}

Get-Content $envFile | ForEach-Object {
    $line = $_.Trim()
    if ($line -eq "" -or $line.StartsWith("#")) { return }
    if ($line -match '^([^=]+)=(.*)$') {
        $name = $matches[1].Trim()
        $value = $matches[2].Trim()
        # Retire une paire de guillemets englobante (simples ou doubles),
        # comme le font la plupart des parseurs .env — sans ça, une valeur
        # ecrite entre guillemets se retrouve avec les guillemets litteraux
        # inclus dans la variable d'environnement.
        if ($value.Length -ge 2 -and (
                ($value.StartsWith('"') -and $value.EndsWith('"')) -or
                ($value.StartsWith("'") -and $value.EndsWith("'"))
            )) {
            $value = $value.Substring(1, $value.Length - 2)
        }
        Set-Item -Path "Env:$name" -Value $value
    }
}

$goCmd = Get-Command go -ErrorAction SilentlyContinue
if ($goCmd) {
    & $goCmd.Source run .
} elseif (Test-Path "C:\Program Files\Go\bin\go.exe") {
    & "C:\Program Files\Go\bin\go.exe" run .
} else {
    Write-Error "go introuvable (ni dans PATH, ni a C:\Program Files\Go\bin\go.exe)"
    exit 1
}
