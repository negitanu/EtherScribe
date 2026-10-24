const {app,dialog}=require('electron');
const fs=require('node:fs');const assert=require('node:assert/strict');
const path = require('node:path');
const root = path.resolve(__dirname, '..');
const temp = fs.mkdtempSync(path.join(require('node:os').tmpdir(), 'etherscribe-smoke-'));
app.setPath('userData', path.join(temp, 'profile'));
app.on('will-quit', () => fs.rmSync(temp, { recursive: true, force: true }));
let inputPath = root+'/core/testdata/junos-ex-els.set';
let cancelSave=false;
dialog.showOpenDialog=async()=>({canceled:false,filePaths:[inputPath]});
dialog.showSaveDialog=async()=>({canceled:cancelSave,filePath:path.join(temp,'report.html')});
app.on('browser-window-created',(_e,w)=>{
 w.webContents.on('did-finish-load',async()=>{
 try {
 const result=await w.webContents.executeJavaScript(`(async()=>{
 const wait=async(fn)=>{for(let i=0;i<100;i++){if(fn())return;await new Promise(r=>setTimeout(r,50))}throw Error('UI timeout: '+document.querySelector('#result').textContent)};
 await wait(()=>!document.querySelector('#preset').disabled);
 if(document.querySelectorAll('#preset option').length!==10)throw Error('preset count');
 document.querySelector('#preset').value='switch-l2';document.querySelector('#preset').dispatchEvent(new Event('change'));
 if(document.querySelectorAll('#sections input:checked').length!==5)throw Error('section selection');
 document.querySelector('#pickConvert').click();await wait(()=>!document.querySelector('#run').disabled);
 document.querySelector('#run').click();await wait(()=>document.querySelector('#resultBadge').textContent==='完了');
 await wait(()=>!document.querySelector('#preview').hidden);
 if(!document.querySelector('#preview').srcdoc.includes('トランク VLAN'))throw Error('report missing switch data');
 return {status:document.querySelector('#engineStatus').textContent,presets:document.querySelectorAll('#preset option').length,summary:document.querySelector('#metrics').textContent};
 })()`);
 console.log('CONVERT',JSON.stringify(result));
 await new Promise(r=>setTimeout(r,400));
 if (process.env.ETHERSCRIBE_SCREENSHOT) fs.writeFileSync(process.env.ETHERSCRIBE_SCREENSHOT,(await w.webContents.capturePage()).toPNG());
 cancelSave=true;
 await w.webContents.executeJavaScript(`document.querySelector('#run').click()`);
 await new Promise(r=>setTimeout(r,200));
 assert.equal(await w.webContents.executeJavaScript(`document.querySelector('#resultBadge').textContent`),'取消');
 assert.equal(await w.webContents.executeJavaScript(`document.querySelector('#preview').hidden`),true);
 cancelSave=false;
 await w.webContents.executeJavaScript(`document.querySelector('#tabCompare').click();document.querySelector('#pickBefore').click()`);
 await new Promise(r=>setTimeout(r,100));
 await w.webContents.executeJavaScript(`document.querySelector('#pickAfter').click()`);
 await new Promise(r=>setTimeout(r,100));
 await w.webContents.executeJavaScript(`document.querySelector('#run').click()`);
 await new Promise(r=>setTimeout(r,300));
 assert.equal(await w.webContents.executeJavaScript(`document.querySelector('#resultBadge').textContent`),'完了');
 assert.equal(await w.webContents.executeJavaScript(`document.querySelector('#preview').srcdoc.includes('比較対象の項目に変更はありません')`),true);
 inputPath=path.join(temp,'missing.conf');
 await w.webContents.executeJavaScript(`document.querySelector('#tabConvert').click();document.querySelector('#pickConvert').click()`);
 await new Promise(r=>setTimeout(r,100));
 await w.webContents.executeJavaScript(`document.querySelector('#run').click()`);
 await new Promise(r=>setTimeout(r,300));
 assert.equal(await w.webContents.executeJavaScript(`document.querySelector('#resultBadge').textContent`),'エラー');
 assert.equal(await w.webContents.executeJavaScript(`document.querySelector('#run').disabled`),false);
 console.log('PASS real Electron IPC: convert, switch preset, preview, save cancellation, compare, error recovery');app.quit();
 }catch(e){console.error(e);app.exit(1)}
 });
});
require(root+'/desktop/dist-electron/main/main.js');
setTimeout(()=>{console.error('Harness timeout');app.exit(1)},15000).unref();
