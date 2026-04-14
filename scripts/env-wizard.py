#!/usr/bin/env python3
import json
import secrets
import sys
from pathlib import Path

ROOT   = Path(__file__).parent.parent
SCHEMA = ROOT / ".env.example.json"
ENV    = ROOT / ".env"
MODE   = sys.argv[1] if len(sys.argv) > 1 else "create"  # create | update

# ── load schema ───────────────────────────────────────────────────────────────

with open(SCHEMA) as f:
    schema = json.load(f)

# ── load existing .env (for update mode) ──────────────────────────────────────

existing = {}
if MODE == "update" and ENV.exists():
    for line in ENV.read_text().splitlines():
        if "=" in line and not line.startswith("#"):
            k, _, v = line.partition("=")
            existing[k.strip()] = v.strip()

# ── helpers ───────────────────────────────────────────────────────────────────

def current(key, default):
    """Return the existing .env value when updating, else the schema default."""
    return existing.get(key, default) if MODE == "update" else default

def profile_value(entry):
    """Return the profile-managed value for entries controlled by the setup profile."""
    key = entry["key"]
    server_val = entry.get("server_docker", "")
    infra_val = entry.get("infra_docker", "")

    if key == "DB_SSLMODE" and (infra_in_docker or server_in_docker):
        return "disable"

    if server_in_docker and server_val:
        return server_val
    if app_uses_docker_hostnames and infra_val:
        return infra_val
    return entry["default"]

def prompt_default(entry):
    """Choose a sensible prompt default for create/update flows."""
    key = entry["key"]
    profile_default = profile_value(entry)

    if MODE != "update":
        return profile_default

    existing_value = existing.get(key)
    if existing_value is None:
        return profile_default

    server_val = entry.get("server_docker", "")
    infra_val = entry.get("infra_docker", "")

    # If the existing value matches an auto-filled value from a different profile,
    # switch the prompt default to the value implied by the newly selected profile.
    if key == "DB_SSLMODE" and existing_value == entry["default"] and profile_default != entry["default"]:
        return profile_default
    if server_val and existing_value == server_val and not server_in_docker:
        return profile_default
    if infra_val and existing_value == infra_val and not app_uses_docker_hostnames:
        return profile_default
    if server_in_docker and server_val:
        return profile_default
    if app_uses_docker_hostnames and infra_val:
        return profile_default

    return existing_value

def ask(prompt, default):
    """Prompt for a free-form value; press Enter to accept the default."""
    label = default if default != "" else "empty"
    val = input(f"  {prompt:<30} [{label}]: ").strip()
    return val if val else default

def ask_choice(prompt, default, choices):
    """Prompt restricted to allowed choices; loops until valid."""
    label = " | ".join(f"{c} (default)" if c == default else c for c in choices)
    opts  = " | ".join(choices)
    while True:
        val = input(f"  {prompt:<30} [{label}]: ").strip()
        val = val if val else default
        if val in choices:
            return val
        print(f"  ✗  Invalid — choose one of: {opts}  (press Enter for '{default}')")

def auto(key, value):
    """Print an auto-assigned value without prompting."""
    print(f"  {key:<30} → {value} (auto)")

def print_entry_help(entry):
    desc = entry.get("desc", "").strip()
    if desc:
        print(f"    {desc}")

def ask_profile():
    print("  Choose setup profile")
    print("    1. Local app + Docker infra (Recommended)")
    print("       Run Go on your machine, connect to PostgreSQL and Redis on localhost.")
    print("    2. Full Docker")
    print("       Run app, PostgreSQL, and Redis in Docker.")
    print("    3. Custom / manual")
    print("       Keep every value editable.")
    print()

    while True:
        choice = input("  Profile                        [1]: ").strip() or "1"
        if choice == "1":
            return {
                "server_in_docker": False,
                "infra_in_docker": True,
                "app_uses_docker_hostnames": False,
                "label": "Local app + Docker infra",
            }
        if choice == "2":
            return {
                "server_in_docker": True,
                "infra_in_docker": True,
                "app_uses_docker_hostnames": True,
                "label": "Full Docker",
            }
        if choice == "3":
            return {
                "server_in_docker": False,
                "infra_in_docker": False,
                "app_uses_docker_hostnames": False,
                "label": "Custom / manual",
            }
        print("  ✗  Invalid — choose 1, 2, or 3.")

def generated_dev_api_key(default):
    if default:
        return default
    return secrets.token_urlsafe(24)

def write_section(lines, section, values, include_infra):
    lines.append(f"# {section['title']}")
    for entry in schema.get(section["key"], []):
        if entry.get("infra_only") and not include_infra:
            continue
        desc = entry.get("desc", "").strip()
        if desc:
            lines.append(f"# {desc}")
        lines.append(f"{entry['key']}={values[entry['key']]}")
    lines.append("")

# ── intro ─────────────────────────────────────────────────────────────────────

print()
if MODE == "update":
    print("  Update .env")
    print("  Press Enter to keep the current value.")
else:
    print("  Create .env")
    print("  Press Enter to accept the recommended default.")
print()

profile = ask_profile()
server_in_docker = profile["server_in_docker"]
infra_in_docker = profile["infra_in_docker"]
app_uses_docker_hostnames = profile["app_uses_docker_hostnames"]

print()
print(f"  Selected profile: {profile['label']}")
if MODE == "create":
    print("  You can still override any prompted value below.")
print()

# ── collect values ────────────────────────────────────────────────────────────

values = {}  # key → value

def process_section(title, section_key):
    print(f"\n  -- {title} {'-' * (44 - len(title))}")
    for entry in schema.get(section_key, []):
        key        = entry["key"]
        default    = prompt_default(entry)
        choices    = entry.get("choices", [])

        if entry.get("infra_only") and not infra_in_docker:
            continue

        server_val = entry.get("server_docker", "")
        infra_val  = entry.get("infra_docker", "")

        if server_in_docker and server_val:
            values[key] = server_val
            auto(key, server_val)
            continue
        if infra_val and app_uses_docker_hostnames:
            values[key] = infra_val
            auto(key, infra_val)
            continue

        print_entry_help(entry)

        # Best practice: non-release modes require a strong dev API key.
        if key == "DEV_API_KEY":
            server_mode = values.get("SERVER_MODE", current("SERVER_MODE", "release"))
            if server_mode != "release":
                default = generated_dev_api_key(default)
                print("    Required outside release mode. A strong key has been prefilled.")

        prompt = key
        if choices:
            values[key] = ask_choice(prompt, default, choices)
        else:
            values[key] = ask(prompt, default)

sections = schema["sections"]
docker_section = schema["docker_section"]

for s in sections:
    process_section(s["title"], s["key"])
if server_in_docker or infra_in_docker:
    process_section(docker_section["title"], docker_section["key"])

# ── write .env ────────────────────────────────────────────────────────────────

write_sections = list(sections)
if server_in_docker or infra_in_docker:
    write_sections.append(docker_section)

lines = []
for s in write_sections:
    write_section(lines, s, values, infra_in_docker)

ENV.write_text("\n".join(lines))

print()
print("  .env updated." if MODE == "update" else "  .env created.")
print(f"  Profile: {profile['label']}")
print()
print("  Next steps")
if server_in_docker:
    print("    1. Run: make docker-dev")
else:
    step = 1
    if infra_in_docker:
        print(f"    {step}. Run: make docker-infra")
        step += 1
    print(f"    {step}. Run: make dev")
print()
