# Remotely Save Support

Obsidian Remotely Save can sync vault files across devices, but the sync scope
should be an explicit choice.

## Safe layouts

### Layout A: Full Obsidian vault sync

- The project directory lives inside an Obsidian vault.
- Remotely Save syncs the vault.
- `AI Memory/` syncs like normal markdown notes.
- Best when the same memory should be available on multiple devices.

### Layout B: Cloud AI Memory folder

- `AI Memory/` is a symlink to a cloud-synced folder.
- Remotely Save syncs the Obsidian vault.
- Codex reads memory through the symlink.
- Best when the code repository and Obsidian vault are separate.

### Layout C: No Obsidian config sync

- Only markdown memory files are synced.
- `.obsidian/` and Remotely Save credentials stay local.
- This is the safest default.

## Safety notes

- Do not generate credential files here.
- Do not commit private plugin config or secrets.
- `.obsidian/` is hidden and may not be synced by default.
- Decide explicitly whether to sync the whole vault.

## Recommended workflow

1. Initialize the normal memory scaffold.
2. Choose the sync layout deliberately.
3. Keep credentials out of `AI Memory/`.
4. Verify what will sync before enabling the plugin.
