import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const viewsDir = path.resolve(__dirname, 'src/views');

function processDirectory(dir) {
  const files = fs.readdirSync(dir);
  for (const file of files) {
    const fullPath = path.join(dir, file);
    if (fs.statSync(fullPath).isDirectory()) {
      processDirectory(fullPath);
    } else if (fullPath.endsWith('.vue')) {
      let content = fs.readFileSync(fullPath, 'utf-8');
      let modified = false;
      
      // Replace style="width: XX%" with style="min-width: XX0px"
      content = content.replace(/style="width:\s*(\d+)%"/g, (match, p1) => {
        modified = true;
        let width = parseInt(p1, 10);
        let minWidth = width * 10;
        // Adjust for small/large values to look better
        if (minWidth < 100) minWidth = 100;
        if (width === 25) minWidth = 200;
        return `style="min-width: ${minWidth}px"`;
      });

      if (modified) {
        fs.writeFileSync(fullPath, content, 'utf-8');
        console.log(`Updated ${fullPath}`);
      }
    }
  }
}

processDirectory(viewsDir);
console.log('Done replacing width percentages with min-width.');
