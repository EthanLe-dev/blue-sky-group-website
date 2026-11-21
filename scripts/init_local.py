"""Create local configuration without overwriting an existing file."""
from pathlib import Path
import secrets
root = Path(__file__).resolve().parent.parent
target = root / '.env'
if target.exists():
    raise SystemExit('.env exists; keep it and fill missing values manually.')
value = (root / '.env.example').read_text(encoding='utf-8')
for key in ['JWT_SIGNING_SECRET','CSRF_SECRET','BOOTSTRAP_ADMIN_PASSWORD']:
    value = value.replace(key+'=\n', key+'='+secrets.token_urlsafe(40)+'\n')
value = value.replace('BOOTSTRAP_ADMIN_EMAIL=\n','BOOTSTRAP_ADMIN_EMAIL=admin@example.com\n')
target.write_text(value,encoding='utf-8')
print('Created .env with random secrets; credentials remain in the ignored file.')
