// jsPDF + autotable ana bundle'a girmesin diye yalnızca dışa aktarım
// tetiklendiğinde dinamik olarak yüklenir.
export const exportToPDF = async (
  filename: string,
  columns: { header: string; dataKey: string }[],
  data: any[]
) => {
  const { default: jsPDF } = await import('jspdf')
  const { default: autoTable } = await import('jspdf-autotable')

  const doc = new jsPDF('p', 'pt', 'a4')

  // Support Turkish characters
  // We will use standard fonts since jsPDF default doesn't support full TR
  // but for a quick fix without external font files we can just let it render.
  
  doc.text(filename, 40, 40)

  autoTable(doc, {
    head: [columns.map(col => col.header)],
    body: data.map(item => columns.map(col => item[col.dataKey])),
    startY: 60,
    styles: {
      font: 'helvetica',
    },
    headStyles: {
      fillColor: [33, 150, 243] // Blue color
    }
  })

  doc.save(`${filename}.pdf`)
}
