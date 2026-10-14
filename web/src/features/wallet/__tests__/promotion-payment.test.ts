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

import { requestPaymentAmount } from '../hooks/use-payment'
import { dispatchSelectedPayment } from '../lib/payment'

describe('promotional payment contracts', () => {
  test('special tier quote preserves the selected tier identifier', async () => {
    let received: unknown
    const calculate = async (request: unknown) => {
      received = request
      return { success: true, data: '2.50' }
    }
    const result = await requestPaymentAmount(
      2.5,
      'alipay',
      {
        regular: calculate,
        stripe: calculate,
        waffo: calculate,
        waffoPancake: calculate,
      },
      { tier_id: 'starter' }
    )
    expect(result).toBe(2.5)
    expect(received).toMatchObject({ amount: 2, tier_id: 'starter' })
  })
  test('bargain order preserves the selected activity identifier', async () => {
    let received: unknown
    await dispatchSelectedPayment(
      { name: 'Alipay', type: 'alipay' },
      2,
      null,
      {
        regular: async (_amount, _type, options) => {
          received = options
          return true
        },
        waffo: async () => false,
        waffoPancake: async () => false,
      },
      { bargain_activity_id: 42 }
    )
    expect(received).toEqual({ bargain_activity_id: 42 })
  })
  test.each(['stripe', 'waffo', 'waffo_pancake'])(
    'promotional orders reject unsupported %s rather than creating a regular order',
    async (type) => {
      let created = false
      const create = async () => {
        created = true
        return true
      }
      const result = await dispatchSelectedPayment(
        { name: type, type },
        2,
        0,
        { regular: create, waffo: create, waffoPancake: create },
        { tier_id: 'starter' }
      )
      expect(result).toBe(false)
      expect(created).toBe(false)
    }
  )
})
