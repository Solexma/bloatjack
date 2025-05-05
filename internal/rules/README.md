# Bloatjack – Embedded Rulebook

This folder (`internal/rules/`) contains the **built‑in optimisation rules** shipped with every release of the Bloatjack CLI.

- Guarantees that Bloatjack works **offline**.
- Ensures every CLI version executes against a deterministic ruleset (reproducibility).
- Acts as the reference implementation that external rulepacks must follow.

---

## Directory layout

```plaintext
internal/rules/
  ├── VERSION              # ISO‑8601 date of the embedded ruleset
  ├── memory.yml           # memory & CPU caps
  ├── volumes.yml          # volume & cache helpers
  ├── language-node.yml    # language‑specific tweaks
  ├── ...
  └── README.md            # you are here
```

## VERSION file

The plain-text `VERSION` file contains a single line in the format `2025.05.05`.
Update this date **every time** you modify any rule in this folder.
`bloatjack --version` prints both the CLI version *and* the rulebook version so users can report bugs precisely.

### Advantages of the VERSION system

1. **Precise ruleset identification**  
   With `bloatjack --version` users see both the binary version (e.g., v0.3.2) and the ruleset version (e.g., 2025-05-05). This allows for exact identification of which rules are in use without asking the user to recompile.

2. **Decoupling code releases from rules**  
   Often you'll only change YAML files (new thresholds, fixes) without touching Go code. By updating `internal/rules/VERSION` and creating a tag like `ruleset-2025-05-10`, you can distribute a minimal new binary (v0.3.3) or even a rules-only package downloadable with `bloatjack rules upgrade`.

3. **CI/CD triggers**  
   In automation workflows, you can specify: "if ANY file in internal/rules/** or VERSION changes, rebuild, generate ruleset-{date}.tar.gz asset and publish". This eliminates forgotten manual pushes.

4. **Reproducibility and debugging**  
   Completes the cycle along with the commit SHA: if you archive a Bloatjack HTML report, inside it displays "ruleset 2025-05-05". Even much later, you can regenerate the same scenario starting from that tag.

5. **Single readable source of truth**  
   A string like "2025.05.05" in a simple file is easier to find and update than a Go constant buried in code. It also doesn't force those who only contribute to rules to modify Go code.

6. **Future compatibility**  
   If you change the schema someday (new keys, different syntax), the loader can read VERSION and implement specific logic: "if ruleset < 2026-01-01, apply migration; otherwise load directly".

### Typical workflow

1. Update a YAML rules file (e.g., memory.yml).
2. Modify `internal/rules/VERSION` with the new date (e.g., 2025-06-01).
3. Commit with message "ruleset 2025-06-01: lower db memory limit".
4. CI detects changes in `internal/rules/**` → build and tag v0.3.4.
5. User runs `brew upgrade` or `bloatjack rules upgrade`.
6. With `bloatjack --version` they now see: Bloatjack v0.3.4 (ruleset 2025-06-01).

## Rule YAML schema

Every rule file is a valid YAML document. Expected top‑level key: `rules:` (array).

```yaml
- id: mem-cap-db@1.2.0       # unique id + semver
  priority: 80               # 0‑100, higher wins conflicts
  match: {kind: db}          # selector on service meta
  if: peak_mem_mb > 800      # JMESPath‑like boolean expression (optional)
  set:                       # patch applied to compose/k8s
    mem_limit: '{peak_mem_mb*1.25}m'
    cpus: '0.25'
  note: 'Cap DB RAM to 125 % of peak and reserve quarter core.'
```

### Core fields

| Field | Required | Description |
|-------|----------|-------------|
|`id`|✅|`<name>@<semver>` — unique across all rule files|
|`priority`|✅|0‑100, resolves multiple rules touching same key|
|`match`|✅|Key/value or wildcard map matched against service metadata|
|`if`| |Boolean expression evaluated on runtime metrics|
|`set`, `ensure_volume`, `set_env`| |Actions that mutate the compose file|
|`action`| |Special verbs: `offload`, `warn`, external `script`|
|`note`| |Human message shown in the HTML report|

## Priority & precedence

1. Embedded rules are loaded first.
2. External rules in `~/.bloatjack/rules.d/` override **only if** they declare an equal or higher `priority`.
3. Rules supplied via `--rulebook` flag override everything.

If two rules edit the same key, the one with the **higher priority number wins**. Tie‑breakers fall back to lexical order of `id`.

## Adding or modifying a rule

1. Place the new rule in the appropriate file *or* create a new `<topic>.yml`.
2. Ensure the `id` is unique; bump the minor version if editing.
3. Run:

   ```bash
   go run ./cmd/rule-lint ./internal/rules
   go test ./internal/rules
   ```

4. Update the `VERSION` file with today's date.
5. Commit in a dedicated PR titled `rulebook: <change summary>`.

## Testing rules manually

```bash
# Launch sample stack
docker compose -f samples/nextjs.yml up -d

# Profile and view suggested patches
bloatjack scan --debug

# Apply patch in dry‑run mode
bloatjack tune --dry-run
```

## Contributing

Pull requests are welcome! Please follow the checklist in `/.github/CONTRIBUTING.md` and add unit tests for new conditions. For bigger changes (new actions, expression predicates) open an issue first to align on direction.

## License

Embedded rules are released under the [MIT license](../LICENSE).
