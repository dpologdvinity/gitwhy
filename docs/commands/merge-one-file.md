# git merge-one-file

Internal: the default program [`git merge-index`](merge-index.md) runs to merge one file (adds, deletes, mode changes, three-way content merge with conflict markers). Use via `git merge-index git-merge-one-file -a`.

**Usage:** `git merge-one-file <orig-blob> <our-blob> <their-blob> <path> <orig-mode> <our-mode> <their-mode>`

## Flags

None. Arguments: base, our, their blob IDs; path; the three modes. Empty blob and mode means that version doesn't exist.
