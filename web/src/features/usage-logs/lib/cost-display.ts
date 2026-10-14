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
import type { LogOtherData } from '../types'

export type UsageCostMarker = 'free'

export interface UsageCostDisplay {
  amount: string
  marker: UsageCostMarker | null
  prefix: string
}

function splitCostPrefix(value: string): { prefix: string; amount: string } {
  const match = value.match(/^([^0-9+\-.,\s]+)(.+)$/)
  if (!match) return { prefix: '', amount: value }
  return { prefix: match[1], amount: match[2] }
}

export function isFreePointsCost(other: LogOtherData | null): boolean {
  if (!other || other.billing_source === 'subscription') return false
  return other.billing_marker === 'free' || Number(other.bonus_quota ?? 0) > 0
}

export function getUsageCostDisplay(
  quotaText: string,
  other: LogOtherData | null
): UsageCostDisplay {
  const quotaDisplay = splitCostPrefix(quotaText)
  if (isFreePointsCost(other)) {
    return {
      amount: quotaDisplay.amount,
      marker: 'free',
      prefix: '',
    }
  }

  return {
    amount: quotaDisplay.amount,
    marker: null,
    prefix: quotaDisplay.prefix,
  }
}
