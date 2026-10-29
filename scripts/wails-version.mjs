import fs from 'node:fs'

const wailsPath = 'wails.json'
const backupPath = '.wails.json.bak'
const [mode, version] = process.argv.slice(2)

function die(message) {
  console.error(message)
  process.exit(1)
}

if (mode === 'set') {
  if (!version) die('usage: node scripts/wails-version.mjs set <version>')
  if (!/^\d+\.\d+\.\d+$/.test(version)) {
    die(`invalid version "${version}"; expected semver like 0.2.0`)
  }
  if (!fs.existsSync(wailsPath)) {
    die('wails.json not found; run this script from the repository root')
  }
  // Recover from an interrupted previous build before taking a fresh backup.
  if (fs.existsSync(backupPath)) {
    fs.copyFileSync(backupPath, wailsPath)
    fs.rmSync(backupPath)
  }
  fs.copyFileSync(wailsPath, backupPath)
  const json = JSON.parse(fs.readFileSync(wailsPath, 'utf8'))
  json.info = json.info || {}
  json.info.productVersion = version
  fs.writeFileSync(wailsPath, `${JSON.stringify(json, null, 2)}\n`)
  console.log(`wails.json productVersion -> ${version}`)
} else if (mode === 'restore') {
  if (fs.existsSync(backupPath)) {
    fs.copyFileSync(backupPath, wailsPath)
    fs.rmSync(backupPath)
    console.log('wails.json restored')
  }
} else {
  die('usage: node scripts/wails-version.mjs set <version> | restore')
}
