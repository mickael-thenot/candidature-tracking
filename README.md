# candidature-tracking

Outil CLI en Go pour suivre mes candidatures.

## Objectif

Suivre l'état de chaque candidature (envoyée, entretien, offre, refus) depuis le terminal.

## Stack

- Go 1.27
- Bibliothèque standard uniquement (`net/http`, `encoding/json`, `crypto/uuid`)

## Roadmap

- [x] Modèle `Candidature` avec validation
- [ ] Interface `Store` et implémentation JSON
- [ ] CLI (`add`, `list`, `update`, `delete`)
- [ ] Tests unitaires