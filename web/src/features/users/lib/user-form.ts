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
import { z } from 'zod'

import {
  type PermissionCatalog,
  type AdminPermissionMatrix,
  normalizeAdminPermissions,
} from '@/lib/admin-permissions'
import { quotaUnitsToDollars } from '@/lib/format'
import { isValidModelRpmMap, serializeModelRpmMap } from '@/lib/rate-limit'
import { ROLE } from '@/lib/roles'

import { DEFAULT_GROUP } from '../constants'
import type { UserFormData, User } from '../types'

// ============================================================================
// Form Schema
// ============================================================================

export const userFormSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  display_name: z.string().optional(),
  password: z.string().optional(),
  role: z.number().optional(),
  quota_dollars: z.number().min(0).optional(),
  group: z.string().optional(),
  remark: z.string().optional(),
  // 免费积分模型 RPM 覆盖值：undefined 表示继承全局默认，0 表示该用户不限
  bonus_quota_model_rpm: z.number().int().min(0).optional(),
  // 按模型的免费积分模型 RPM 覆盖表，表单里是 JSON 文本，空串表示不覆盖
  bonus_quota_model_rpm_by_model: z
    .string()
    .optional()
    .refine(isValidModelRpmMap, {
      message: 'Invalid JSON format or values out of allowed range',
    }),
  admin_permissions: z
    .record(z.string(), z.record(z.string(), z.boolean()))
    .optional(),
})

export type UserFormValues = z.infer<typeof userFormSchema>

// ============================================================================
// Form Defaults
// ============================================================================

export const USER_FORM_DEFAULT_VALUES: UserFormValues = {
  username: '',
  display_name: '',
  password: '',
  role: 1, // Default to common user
  quota_dollars: 0,
  group: DEFAULT_GROUP,
  remark: '',
  bonus_quota_model_rpm: undefined,
  bonus_quota_model_rpm_by_model: '',
  // Filled against the backend catalog at render time; see UsersMutateDrawer.
  admin_permissions: {},
}

// ============================================================================
// Form Data Transformation
// ============================================================================

/**
 * Transform form data to API payload
 */
export function transformFormDataToPayload(
  data: UserFormValues,
  userId?: number,
  catalog?: PermissionCatalog
): UserFormData & { id?: number } {
  const payload: UserFormData & { id?: number } = {
    username: data.username,
    display_name: data.display_name || data.username,
    password: data.password || undefined,
  }

  const role = userId === undefined ? data.role || 1 : (data.role ?? 0)

  // Only send the permission matrix when the target is an admin and the catalog
  // is available; without the catalog we cannot build a full matrix, so we omit
  // the field (the backend then leaves existing permissions untouched).
  if (role >= ROLE.ADMIN && catalog) {
    payload.admin_permissions = normalizeAdminPermissions(
      data.admin_permissions as AdminPermissionMatrix | undefined,
      catalog
    )
  }

  // For create: only send required fields
  if (userId === undefined) {
    payload.role = role
  } else {
    // For update: quota is adjusted atomically via /api/user/manage, not sent here
    payload.group = data.group
    payload.remark = data.remark || undefined
    payload.id = userId
  }

  return payload
}

/**
 * 从用户 setting JSON 中读出免费积分模型 RPM 覆盖值。
 * 返回 undefined 表示该用户没有覆盖，运行时回落到全局默认设置。
 */
function parseBonusQuotaModelRpm(setting?: string): number | undefined {
  if (!setting) return undefined
  try {
    const value = JSON.parse(setting)?.bonus_quota_model_rpm
    return typeof value === 'number' ? value : undefined
  } catch {
    return undefined
  }
}

/**
 * 把用户 setting JSON 里的按模型 RPM 覆盖表转成表单用的 JSON 文本。
 */
function parseBonusQuotaModelRpmByModel(setting?: string): string {
  if (!setting) return ''
  try {
    return serializeModelRpmMap(
      JSON.parse(setting)?.bonus_quota_model_rpm_by_model
    )
  } catch {
    return ''
  }
}

/**
 * Transform user data to form defaults. The admin permission matrix is passed
 * through as-is (the backend already returns a full matrix); it is filled against
 * the catalog at render time in UsersMutateDrawer.
 */
export function transformUserToFormDefaults(user: User): UserFormValues {
  return {
    username: user.username,
    display_name: user.display_name,
    password: '',
    role: user.role,
    quota_dollars: quotaUnitsToDollars(user.quota),
    group: user.group || DEFAULT_GROUP,
    remark: user.remark || '',
    bonus_quota_model_rpm: parseBonusQuotaModelRpm(user.setting),
    bonus_quota_model_rpm_by_model: parseBonusQuotaModelRpmByModel(
      user.setting
    ),
    admin_permissions: user.admin_permissions ?? {},
  }
}
