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
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { StaticDataTable } from '@/components/data-table'
import { EmptyState } from '@/components/empty-state'
import { ErrorState } from '@/components/error-state'
import { SectionPageLayout } from '@/components/layout'
import { LoadingState } from '@/components/loading-state'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import dayjs from '@/lib/dayjs'
import { formatQuota } from '@/lib/format'
import { requireServerSuccess } from '@/lib/server-error-message'

import { getAdminAffiliateCommissions } from './api'
import type { AffiliateCommissionAdminFilters } from './types'

interface DraftFilters {
  inviterId: string
  inviteeId: string
  tradeNo: string
  provider: string
  status: string
  startTime: string
  endTime: string
}

const emptyDraftFilters: DraftFilters = {
  inviterId: '',
  inviteeId: '',
  tradeNo: '',
  provider: '',
  status: 'all',
  startTime: '',
  endTime: '',
}

const emptyFilters: AffiliateCommissionAdminFilters = {
  inviterId: '',
  inviteeId: '',
  tradeNo: '',
  provider: '',
  status: '',
  startTimestamp: '',
  endTimestamp: '',
}

const SKIPPED_REASONS: Record<string, string> = {
  no_inviter: 'No inviter',
  inviter_deleted: 'Inviter deleted',
  missing_commission_base: 'Missing commission base',
  rounded_zero: 'Rounded to zero',
}

export function AffiliateCommissions() {
  const { t } = useTranslation()
  const [draft, setDraft] = useState<DraftFilters>(emptyDraftFilters)
  const [filters, setFilters] =
    useState<AffiliateCommissionAdminFilters>(emptyFilters)
  const [page, setPage] = useState(1)
  const pageSize = 20
  const query = useQuery({
    queryKey: ['admin-affiliate-commissions', page, pageSize, filters],
    queryFn: async () => {
      const response = await getAdminAffiliateCommissions(
        page,
        pageSize,
        filters
      )
      requireServerSuccess(response)
      return (
        response.data ?? { page: 1, page_size: pageSize, total: 0, items: [] }
      )
    },
    placeholderData: (previous) => previous,
  })
  const data = query.data

  const setFilter = (key: keyof DraftFilters, value: string) => {
    setDraft((current) => ({ ...current, [key]: value }))
  }

  const handleApply = () => {
    let startTimestamp = ''
    let endTimestamp = ''
    if (draft.startTime) {
      const s = Math.floor(new Date(draft.startTime).getTime() / 1000)
      if (!Number.isNaN(s) && s > 0) startTimestamp = String(s)
    }
    if (draft.endTime) {
      const e = Math.floor(new Date(draft.endTime).getTime() / 1000)
      if (!Number.isNaN(e) && e > 0) endTimestamp = String(e)
    }
    if (
      startTimestamp &&
      endTimestamp &&
      Number(startTimestamp) > Number(endTimestamp)
    ) {
      toast.error(t('Start time cannot be later than end time'))
      return
    }
    setPage(1)
    setFilters({
      inviterId: draft.inviterId.trim(),
      inviteeId: draft.inviteeId.trim(),
      tradeNo: draft.tradeNo.trim(),
      provider: draft.provider.trim(),
      status: draft.status === 'all' ? '' : draft.status,
      startTimestamp,
      endTimestamp,
    })
  }

  const handleReset = () => {
    setDraft(emptyDraftFilters)
    setFilters(emptyFilters)
    setPage(1)
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        {t('Top-up commissions')}
      </SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='flex flex-col gap-4'>
          <FieldGroup className='grid gap-3 md:grid-cols-4'>
            <Field>
              <FieldLabel htmlFor='commission-inviterId'>
                {t('Inviter ID')}
              </FieldLabel>
              <Input
                type='number'
                min={1}
                id='commission-inviterId'
                value={draft.inviterId}
                onChange={(event) => setFilter('inviterId', event.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor='commission-inviteeId'>
                {t('Invitee ID')}
              </FieldLabel>
              <Input
                type='number'
                min={1}
                id='commission-inviteeId'
                value={draft.inviteeId}
                onChange={(event) => setFilter('inviteeId', event.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor='commission-tradeNo'>
                {t('Trade number')}
              </FieldLabel>
              <Input
                id='commission-tradeNo'
                value={draft.tradeNo}
                onChange={(event) => setFilter('tradeNo', event.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor='commission-provider'>
                {t('Provider')}
              </FieldLabel>
              <Input
                id='commission-provider'
                value={draft.provider}
                onChange={(event) => setFilter('provider', event.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel>{t('Status')}</FieldLabel>
              <Select
                value={draft.status}
                onValueChange={(value) => setFilter('status', value ?? 'all')}
              >
                <SelectTrigger className='w-full'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent alignItemWithTrigger={false}>
                  <SelectGroup>
                    <SelectItem value='all'>{t('All statuses')}</SelectItem>
                    <SelectItem value='awarded'>{t('Awarded')}</SelectItem>
                    <SelectItem value='skipped'>{t('Skipped')}</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel htmlFor='commission-startTime'>
                {t('Start time')}
              </FieldLabel>
              <Input
                type='datetime-local'
                id='commission-startTime'
                value={draft.startTime}
                onChange={(event) => setFilter('startTime', event.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor='commission-endTime'>
                {t('End time')}
              </FieldLabel>
              <Input
                type='datetime-local'
                id='commission-endTime'
                value={draft.endTime}
                onChange={(event) => setFilter('endTime', event.target.value)}
              />
            </Field>
            <Field className='justify-end'>
              <FieldLabel className='sr-only'>{t('Filter actions')}</FieldLabel>
              <div className='flex gap-2'>
                <Button onClick={handleApply}>{t('Apply filters')}</Button>
                <Button variant='outline' onClick={handleReset}>
                  {t('Reset')}
                </Button>
              </div>
            </Field>
          </FieldGroup>

          {query.isError && <ErrorState onRetry={() => query.refetch()} />}
          {query.isLoading && <LoadingState />}
          {!query.isError &&
            !query.isLoading &&
            (data?.items.length ? (
              <StaticDataTable>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('Trade number')}</TableHead>
                    <TableHead>{t('Inviter')}</TableHead>
                    <TableHead>{t('Invitee')}</TableHead>
                    <TableHead>{t('Payment')}</TableHead>
                    <TableHead>{t('Provider')}</TableHead>
                    <TableHead>{t('Rate')}</TableHead>
                    <TableHead>{t('Commission')}</TableHead>
                    <TableHead>{t('Status')}</TableHead>
                    <TableHead>{t('Skipped reason')}</TableHead>
                    <TableHead>{t('Time')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.items.map((record) => (
                    <TableRow key={record.id}>
                      <TableCell>{record.trade_no}</TableCell>
                      <TableCell>{record.inviter_id}</TableCell>
                      <TableCell>{record.invitee_id}</TableCell>
                      <TableCell>
                        {record.actual_paid_amount} {record.currency}
                      </TableCell>
                      <TableCell>{record.payment_provider}</TableCell>
                      <TableCell>
                        {(record.rate_basis_points / 100).toFixed(2)}%
                      </TableCell>
                      <TableCell>
                        {formatQuota(record.commission_quota)}
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={
                            record.status === 'awarded'
                              ? 'default'
                              : 'secondary'
                          }
                        >
                          {t(
                            record.status === 'awarded' ? 'Awarded' : 'Skipped'
                          )}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {record.skipped_reason
                          ? t(
                              SKIPPED_REASONS[record.skipped_reason] ||
                                record.skipped_reason
                            )
                          : '-'}
                      </TableCell>
                      <TableCell>
                        {dayjs
                          .unix(record.create_time)
                          .format('YYYY-MM-DD HH:mm')}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </StaticDataTable>
            ) : (
              <EmptyState
                title={t('No commission records found')}
                description={t('Try adjusting the commission filters.')}
              />
            ))}
          <div className='flex justify-end gap-2'>
            <Button
              variant='outline'
              disabled={page <= 1}
              onClick={() => setPage((value) => value - 1)}
            >
              {t('Previous')}
            </Button>
            <Button
              variant='outline'
              disabled={!data || page * pageSize >= data.total}
              onClick={() => setPage((value) => value + 1)}
            >
              {t('Next')}
            </Button>
          </div>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
