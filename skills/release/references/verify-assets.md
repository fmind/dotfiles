# Verify Release Assets

Download the expected inventory into a fresh temporary directory, verify the repository's checksum manifest, then verify integrity and provenance where the release publishes attestations.

```bash
release_dir=$(mktemp -d)
gh release download "$tag" -R "$repo" --dir "$release_dir"
gh release verify "$tag" -R "$repo"
gh release verify-asset "$tag" "$release_dir/checksums.txt" -R "$repo"
(cd "$release_dir" && shasum -a 256 -c checksums.txt)
asset=project_vX.Y.Z_linux_amd64.tar.gz
signer_workflow="$repo/.github/workflows/release.yml"
gh release verify-asset "$tag" "$release_dir/$asset" -R "$repo"
gh attestation verify "$release_dir/$asset" --repo "$repo" --signer-workflow "$signer_workflow" --source-ref "refs/tags/$tag" --source-digest "$release_sha"
```

- `gh release verify` and `gh release verify-asset` apply only to immutable releases with release attestations; generated source archives cannot be verified as uploaded assets.
- Verify the checksum manifest itself before trusting its entries.
- Run `gh attestation verify` when separate build provenance exists, with the expected signer workflow.
- Missing assets, checksums, attestations, or signer identity are a failed proof, not permission to regenerate or replace them.
