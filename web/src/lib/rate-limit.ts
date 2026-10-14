/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
/**
 * 「模型名 -> 每分钟请求数上限」JSON 文本的格式校验。
 * 空值合法（表示未配置）；非空时必须是 模型名 -> 非负整数 的 JSON 对象。
 * 模型名不做名单校验：免费积分模型名单会变动，硬校验会让配置在模型被移出名单时变成非法。
 */
export function isValidModelRpmMap(value: string | undefined): boolean {
  if (!value || value.trim() === '') return true
  try {
    const parsed = JSON.parse(value)
    if (
      typeof parsed !== 'object' ||
      parsed === null ||
      Array.isArray(parsed)
    ) {
      return false
    }
    return Object.entries(parsed).every(
      ([model, limit]) =>
        model.trim() !== '' &&
        typeof limit === 'number' &&
        Number.isInteger(limit) &&
        limit >= 0
    )
  } catch {
    return false
  }
}

/**
 * 把后端返回的 setting 对象里的按模型 RPM 表转成表单用的 JSON 文本。
 * 未配置时返回空字符串，表示不覆盖（继承全局）。
 */
export function serializeModelRpmMap(
  value: Record<string, number> | undefined
): string {
  if (!value || Object.keys(value).length === 0) return ''
  return JSON.stringify(value)
}
