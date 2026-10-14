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
export interface AdminAffiliateCommissionRecord {
  id: number
  topup_id: number
  trade_no: string
  inviter_id: number
  invitee_id: number
  payment_provider: string
  payment_method: string
  actual_paid_amount: string
  currency: string
  commission_base_quota: number
  rate_basis_points: number
  commission_quota: number
  status: 'awarded' | 'skipped'
  skipped_reason?: string
  create_time: number
}

export interface AffiliateCommissionAdminFilters {
  inviterId: string
  inviteeId: string
  tradeNo: string
  provider: string
  status: string
  startTimestamp: string
  endTimestamp: string
}

export interface AdminAffiliateCommissionPage {
  page: number
  page_size: number
  total: number
  items: AdminAffiliateCommissionRecord[]
}
