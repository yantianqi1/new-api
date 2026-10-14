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
import { afterEach, describe, expect, test, vi } from 'vitest'

import { api } from '@/lib/api'

import { updateUserBonusQuotaModelRpm } from '../../api'
import { userSchema } from '../../types'
import {
  transformUserToFormDefaults,
  userFormSchema,
  USER_FORM_DEFAULT_VALUES,
} from '../user-form'

afterEach(() => vi.restoreAllMocks())
describe('free-credit model RPM overrides', () => {
  test('a zero override and per-model overrides survive loading a user', () => {
    const user = userSchema.parse({
      id: 7,
      username: 'limited',
      display_name: 'Limited',
      role: 1,
      status: 1,
      quota: 0,
      used_quota: 0,
      request_count: 0,
      group: 'default',
      setting: JSON.stringify({
        bonus_quota_model_rpm: 0,
        bonus_quota_model_rpm_by_model: { free: 0, frame: 3 },
      }),
    })
    const values = transformUserToFormDefaults(user)
    expect(values.bonus_quota_model_rpm).toBe(0)
    expect(JSON.parse(values.bonus_quota_model_rpm_by_model ?? '{}')).toEqual({
      free: 0,
      frame: 3,
    })
  })
  test('clearing the overrides explicitly sends null to both authoritative fields', async () => {
    const put = vi
      .spyOn(api, 'put')
      .mockResolvedValue({ data: { success: true } })
    await updateUserBonusQuotaModelRpm(7, null, null)
    expect(put).toHaveBeenCalledWith('/api/user/7/bonus_quota_model_rpm', {
      bonus_quota_model_rpm: null,
      bonus_quota_model_rpm_by_model: null,
    })
  })
  test('fractional per-minute limits are rejected before saving', () => {
    expect(
      userFormSchema.safeParse({
        ...USER_FORM_DEFAULT_VALUES,
        username: 'limited',
        bonus_quota_model_rpm: 1.5,
      }).success
    ).toBe(false)
    expect(
      userFormSchema.safeParse({
        ...USER_FORM_DEFAULT_VALUES,
        username: 'limited',
        bonus_quota_model_rpm_by_model: '{"free":1.5}',
      }).success
    ).toBe(false)
  })
})
