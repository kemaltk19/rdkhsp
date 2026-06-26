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
  
  // Replace <Button ... />
  // We need a regex that matches <Button ... /> or <Button ... >
  const buttonRegex = /<Button([^>]+)>/g
  
  content = content.replace(buttonRegex, (match, attrs) => {
    let newAttrs = attrs
    
    // Check if it's an icon-only button (has 'icon=' but no 'label=')
    const hasLabel = /label=/.test(attrs)
    const hasIcon = /icon=/.test(attrs)
    
    // Remove existing styling props/classes if we are standardizing
    // We'll leave existing classes to avoid breaking margins like 'mr-2'
    // But let's add the boolean props if missing
    
    if (hasLabel) {
      if (!/\boutlined\b/.test(newAttrs)) {
        newAttrs += ' outlined'
        changed = true
      }
    } else if (hasIcon) {
      if (!/\btext\b/.test(newAttrs)) {
        newAttrs += ' text'
        changed = true
      }
      if (!/\brounded\b/.test(newAttrs)) {
        newAttrs += ' rounded'
        changed = true
      }
    } else {
      // default text button with content inside <Button>...</Button>
      if (!/\boutlined\b/.test(newAttrs)) {
        newAttrs += ' outlined'
        changed = true
      }
    }
    
    return `<Button${newAttrs}>`
  })
  
  if (changed) {
    fs.writeFileSync(filePath, content, 'utf8')
    console.log(`Updated: ${filePath}`)
  }
}

processDirectory(srcDir)
