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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { requireServerSuccess } from '@/lib/server-error-message'

import {
  getAffiliateCommissionRecords,
  getAffiliateCommissionSummary,
  getAffiliateCommissionTransfers,
  transferAffiliateCommission,
} from '../api'

export function useAffiliateCommission(
  recordPage: number,
  transferPage: number
) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const pageSize = 10

  const summary = useQuery({
    queryKey: ['affiliate-commission-summary'],
    queryFn: async () =>
      requireServerSuccess(await getAffiliateCommissionSummary()),
  })
  const records = useQuery({
    queryKey: ['affiliate-commission-records', recordPage, pageSize],
    queryFn: async () =>
      requireServerSuccess(
        await getAffiliateCommissionRecords(recordPage, pageSize)
      ),
    placeholderData: (previous) => previous,
  })
  const transfers = useQuery({
    queryKey: ['affiliate-commission-transfers', transferPage, pageSize],
    queryFn: async () =>
      requireServerSuccess(
        await getAffiliateCommissionTransfers(transferPage, pageSize)
      ),
    placeholderData: (previous) => previous,
  })
  const transfer = useMutation({
    mutationFn: async (
      request: Parameters<typeof transferAffiliateCommission>[0]
    ) => requireServerSuccess(await transferAffiliateCommission(request)),
    onSuccess: async (response) => {
      if (!response.success) {
        toast.error(response.message || t('Commission transfer failed'))
        return
      }
      toast.success(t('Commission transferred to paid quota'))
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['affiliate-commission-summary'],
        }),
        queryClient.invalidateQueries({
          queryKey: ['affiliate-commission-transfers'],
        }),
      ])
    },
  })

  return { summary, records, transfers, transfer, pageSize }
}
