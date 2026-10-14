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
import { describe, expect, test } from 'vitest'

import {
  promotionsSchema,
  PROMOTIONS_DEFAULT_VALUES,
} from '../promotions-schema'

describe('promotion settings', () => {
  test('supports stocked tiers and the production bargain option keys', () => {
    const values = promotionsSchema.parse({
      ...PROMOTIONS_DEFAULT_VALUES,
      payment_setting_special_tiers: JSON.stringify([
        {
          id: 'starter',
          name: 'Starter',
          price: 1,
          quota: 500000,
          total_limit: 10,
          user_limit: 1,
          description: '',
          enabled: true,
          sort_order: 0,
          show_stock: false,
        },
      ]),
      bargain_setting_quota: 1500000,
    })
    expect(
      JSON.parse(values['payment_setting_special_tiers'])[0].show_stock
    ).toBe(false)
    expect(values['bargain_setting_quota']).toBe(1500000)
  })
  test.each([
    ['bargain_setting_max_slashes', 1.5],
    ['bargain_setting_quota', 0],
    ['bargain_setting_min_price', 0],
    ['bargain_setting_user_limit', -1],
  ])('rejects invalid %s', (key, value) => {
    expect(
      promotionsSchema.safeParse({ ...PROMOTIONS_DEFAULT_VALUES, [key]: value })
        .success
    ).toBe(false)
  })
  test('rejects an invalid tier instead of submitting it', () => {
    expect(
      promotionsSchema.safeParse({
        ...PROMOTIONS_DEFAULT_VALUES,
        payment_setting_special_tiers: '[{"name":"Bad","price":0,"quota":0}]',
      }).success
    ).toBe(false)
  })
})
