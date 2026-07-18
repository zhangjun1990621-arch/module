/**
 * 铝厂数据导出工具 — 生成 CSV（兼容 Excel 打开）
 * 参考原铝厂项目的 XLSX 导出能力，简化为 CSV 方案（无需额外依赖）
 */

/** 将数据导出为 CSV 文件 */
export function exportToCSV(
  filename: string,
  headers: { label: string; prop: string }[],
  data: Record<string, any>[]
): void {
  // BOM 头确保 Excel 正确识别 UTF-8
  const BOM = '\uFEFF'
  const headerRow = headers.map((h) => `"${h.label}"`).join(',')
  const dataRows = data.map((row) =>
    headers.map((h) => {
      const val = row[h.prop]
      if (val === null || val === undefined) return '""'
      const str = String(val).replace(/"/g, '""')
      return `"${str}"`
    }).join(',')
  )
  const csv = BOM + headerRow + '\n' + dataRows.join('\n')

  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${filename}_${formatDate(new Date())}.csv`
  link.click()
  URL.revokeObjectURL(url)
}

function formatDate(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}${p(d.getMonth() + 1)}${p(d.getDate())}_${p(d.getHours())}${p(d.getMinutes())}`
}
