import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const srcDir = path.join(__dirname, 'src')

function processDirectory(dir) {
  const files = fs.readdirSync(dir)
  
  for (const file of files) {
    const fullPath = path.join(dir, file)
    const stat = fs.statSync(fullPath)
    
    if (stat.isDirectory()) {
      processDirectory(fullPath)
    } else if (file.endsWith('.vue')) {
      processFile(fullPath)
    }
  }
}

function processFile(filePath) {
  let content = fs.readFileSync(filePath, 'utf8')
  let changed = false
  
  const replacements = [
    { from: /\/ outlined>/g, to: 'outlined />' },
    { from: /\/ text rounded>/g, to: 'text rounded />' },
    { from: /\/ text>/g, to: 'text />' },
    { from: /\/ rounded>/g, to: 'rounded />' },
    // Handle cases where there were spaces
    { from: /\/\s+outlined>/g, to: 'outlined />' },
    { from: /\/\s+text rounded>/g, to: 'text rounded />' },
  ]
  
  for (const req of replacements) {
    if (req.from.test(content)) {
      content = content.replace(req.from, req.to)
      changed = true
    }
  }
  
  if (changed) {
    fs.writeFileSync(filePath, content, 'utf8')
    console.log(`Fixed: ${filePath}`)
  }
}

processDirectory(srcDir)
