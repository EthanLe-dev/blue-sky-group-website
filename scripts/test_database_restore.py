"""Local-only restore drill in two uniquely named disposable test databases."""
from pathlib import Path
import subprocess, uuid
root=Path(__file__).resolve().parent.parent
values=dict(line.split('=',1) for line in (root/'.env').read_text().splitlines() if line and not line.startswith('#') and '=' in line)
user=values['POSTGRES_USER']
names=['restore_'+uuid.uuid4().hex+'_test' for _ in range(2)]
base=['docker','compose','exec','-T','postgres']
def run(args,data=None):
 return subprocess.run(base+args,cwd=root,input=data,stdout=subprocess.PIPE,stderr=subprocess.PIPE,check=True).stdout
created=[]
try:
 for name in names:
  assert name.startswith('restore_') and name.endswith('_test')
  run(['createdb','-U',user,name]);created.append(name)
 for path in sorted((root/'backend/internal/platform/migrations').glob('*.sql')):
  run(['psql','-U',user,'-v','ON_ERROR_STOP=1','-d',names[0]],path.read_bytes())
 run(['psql','-U',user,'-v','ON_ERROR_STOP=1','-d',names[0],'-c',"UPDATE site_settings SET version=2 WHERE key='booking_defaults'"])
 dump=run(['pg_dump','-U',user,'-Fc','--no-owner',names[0]])
 run(['pg_restore','-U',user,'--no-owner','--exit-on-error','-d',names[1]],dump)
 result=run(['psql','-U',user,'-d',names[1],'-Atqc',"SELECT count(*) FROM roles; SELECT version FROM site_settings WHERE key='booking_defaults'"])
 assert result.decode().split()==['5','2']
 print('PASS: migrations, seed, custom setting and restore verified in isolated test databases.')
finally:
 for name in created:
  assert name.startswith('restore_') and name.endswith('_test')
  run(['dropdb','-U',user,name])
