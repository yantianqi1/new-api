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
import assert from 'node:assert/strict'
import test from 'node:test'

import { getUsageCostDisplay } from './cost-display.ts'

test('keeps the currency prefix for paid usage costs', () => {
  assert.deepEqual(
    getUsageCostDisplay('$0.030000', { billing_source: 'wallet' }),
    {
      amount: '0.030000',
      marker: null,
      prefix: '$',
    }
  )
})

test('uses the free marker instead of currency prefix for free points costs', () => {
  assert.deepEqual(
    getUsageCostDisplay('$0.030000', {
      billing_source: 'wallet',
      bonus_quota: 15000,
      paid_quota: 0,
    }),
    {
      amount: '0.030000',
      marker: 'free',
      prefix: '',
    }
  )
})

test('supports the explicit backend free billing marker', () => {
  assert.deepEqual(
    getUsageCostDisplay('$0.030000', {
      billing_source: 'wallet',
      billing_marker: 'free',
    }),
    {
      amount: '0.030000',
      marker: 'free',
      prefix: '',
    }
  )
})
