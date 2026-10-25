// Ad-hoc signing only: no developer certificate, credentials or notarization.
const path = require('node:path');
const { execFileSync } = require('node:child_process');
module.exports = async context => {
  if (context.electronPlatformName !== 'darwin') return;
  const app = path.join(context.appOutDir, `${context.packager.appInfo.productFilename}.app`);
  const engine = path.join(app, 'Contents', 'Resources', 'engine', 'etherscribe');
  execFileSync('/usr/bin/codesign', ['--force', '--sign', '-', '--timestamp=none', engine], { stdio: 'inherit' });
  execFileSync('/usr/bin/codesign', ['--force', '--deep', '--sign', '-', '--timestamp=none', '--preserve-metadata=entitlements', app], { stdio: 'inherit' });
  execFileSync('/usr/bin/codesign', ['--verify', '--deep', '--strict', app], { stdio: 'inherit' });
};
