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
import { BadgeDollarSign } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { StaticDataTable } from '@/components/data-table'
import { EmptyState } from '@/components/empty-state'
import { ErrorState } from '@/components/error-state'
import { LoadingState } from '@/components/loading-state'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Field, FieldDescription, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import {
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import dayjs from '@/lib/dayjs'
import { formatQuota } from '@/lib/format'
import { handleServerError } from '@/lib/handle-server-error'

import { QUOTA_PER_DOLLAR } from '../constants'
import { useAffiliateCommission } from '../hooks/use-affiliate-commission'

interface AffiliateCommissionCardProps {
  onPaidQuotaChanged: () => void | Promise<void>
}

export function AffiliateCommissionCard(props: AffiliateCommissionCardProps) {
  const { t } = useTranslation()
  const [amount, setAmount] = useState('')
  const [recordPage, setRecordPage] = useState(1)
  const [transferPage, setTransferPage] = useState(1)
  const commission = useAffiliateCommission(recordPage, transferPage)
  const summary = commission.summary.data?.data
  const records = commission.records.data?.data
  const transfers = commission.transfers.data?.data

  const transfer = async (all: boolean) => {
    const parsed = Number(amount)
    try {
      const response = await commission.transfer.mutateAsync(
        all ? { all: true } : { quota: parsed }
      )
      if (response.success) {
        setAmount('')
        await props.onPaidQuotaChanged()
      }
    } catch (error) {
      handleServerError(error)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <BadgeDollarSign aria-hidden='true' />
          {t('Wallet top-up commission')}
        </CardTitle>
        <CardDescription>
          {t(
            'Commission from eligible invitees is kept separately and transfers only to paid quota.'
          )}
        </CardDescription>
      </CardHeader>
      <CardContent className='flex flex-col gap-5'>
        {commission.summary.isError && (
          <ErrorState onRetry={() => commission.summary.refetch()} />
        )}
        {!commission.summary.isError &&
          (commission.summary.isLoading ? (
            <div className='grid gap-3 sm:grid-cols-3'>
              {[0, 1, 2].map((item) => (
                <Skeleton key={item} className='h-16' />
              ))}
            </div>
          ) : (
            <div className='grid gap-3 sm:grid-cols-3'>
              <Metric
                label={t('Available paid commission')}
                value={formatQuota(summary?.pending_quota ?? 0)}
              />
              <Metric
                label={t('Lifetime paid commission')}
                value={formatQuota(summary?.history_quota ?? 0)}
              />
              <Metric
                label={t('Eligible invitees')}
                value={String(summary?.eligible_invitee_count ?? 0)}
              />
            </div>
          ))}

        <FieldGroupTransfer
          amount={amount}
          pendingQuota={summary?.pending_quota ?? 0}
          loading={
            commission.transfer.isPending ||
            commission.summary.isLoading ||
            commission.summary.isError
          }
          onAmountChange={setAmount}
          onPartial={() => transfer(false)}
          onAll={() => transfer(true)}
          t={t}
        />

        <Tabs defaultValue='commissions'>
          <TabsList>
            <TabsTrigger value='commissions'>
              {t('Commission details')}
            </TabsTrigger>
            <TabsTrigger value='transfers'>{t('Transfer history')}</TabsTrigger>
          </TabsList>
          <TabsContent value='commissions'>
            {commission.records.isError && (
              <ErrorState onRetry={() => commission.records.refetch()} />
            )}
            {commission.records.isLoading && <LoadingState />}
            {!commission.records.isError &&
              !commission.records.isLoading &&
              (records?.items.length ? (
                <StaticDataTable>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('Invitee')}</TableHead>
                      <TableHead>{t('Payment')}</TableHead>
                      <TableHead>{t('Provider')}</TableHead>
                      <TableHead>{t('Rate')}</TableHead>
                      <TableHead>{t('Commission')}</TableHead>
                      <TableHead>{t('Status')}</TableHead>
                      <TableHead>{t('Time')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {records.items.map((record) => (
                      <TableRow key={record.id}>
                        <TableCell>{record.masked_invitee}</TableCell>
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
                              record.status === 'awarded'
                                ? 'Awarded'
                                : 'Skipped'
                            )}
                          </Badge>
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
                <CommissionEmpty
                  title={t('No commission records yet')}
                  description={t('Eligible wallet top-ups will appear here.')}
                />
              ))}
            <Pager
              page={recordPage}
              total={records?.total ?? 0}
              pageSize={commission.pageSize}
              onPageChange={setRecordPage}
              t={t}
            />
          </TabsContent>
          <TabsContent value='transfers'>
            {commission.transfers.isError && (
              <ErrorState onRetry={() => commission.transfers.refetch()} />
            )}
            {commission.transfers.isLoading && <LoadingState />}
            {!commission.transfers.isError &&
              !commission.transfers.isLoading &&
              (transfers?.items.length ? (
                <StaticDataTable>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('Transferred')}</TableHead>
                      <TableHead>{t('Pending after')}</TableHead>
                      <TableHead>{t('Paid quota after')}</TableHead>
                      <TableHead>{t('Time')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {transfers.items.map((record) => (
                      <TableRow key={record.id}>
                        <TableCell>{formatQuota(record.quota)}</TableCell>
                        <TableCell>
                          {formatQuota(record.pending_quota_after)}
                        </TableCell>
                        <TableCell>
                          {formatQuota(record.paid_quota_after)}
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
                <CommissionEmpty
                  title={t('No commission transfers yet')}
                  description={t('Transfers into paid quota will appear here.')}
                />
              ))}
            <Pager
              page={transferPage}
              total={transfers?.total ?? 0}
              pageSize={commission.pageSize}
              onPageChange={setTransferPage}
              t={t}
            />
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}

function Metric(props: { label: string; value: string }) {
  return (
    <div className='bg-muted/40 rounded-lg p-3'>
      <div className='text-muted-foreground text-xs'>{props.label}</div>
      <div className='mt-1 text-lg font-semibold tabular-nums'>
        {props.value}
      </div>
    </div>
  )
}

function FieldGroupTransfer(props: {
  amount: string
  pendingQuota: number
  loading: boolean
  onAmountChange: (value: string) => void
  onPartial: () => void
  onAll: () => void
  t: (key: string) => string
}) {
  const parsed = Number(props.amount)
  const invalid =
    props.amount !== '' &&
    (!Number.isInteger(parsed) ||
      parsed < QUOTA_PER_DOLLAR ||
      parsed > props.pendingQuota)
  return (
    <Field data-invalid={invalid || undefined}>
      <FieldLabel htmlFor='affiliate-commission-transfer'>
        {props.t('Partial transfer amount')}
      </FieldLabel>
      <div className='flex flex-col gap-2 sm:flex-row'>
        <Input
          id='affiliate-commission-transfer'
          type='number'
          min={QUOTA_PER_DOLLAR}
          max={props.pendingQuota}
          step={1}
          value={props.amount}
          onChange={(event) => props.onAmountChange(event.target.value)}
          aria-invalid={invalid}
        />
        <Button
          onClick={props.onPartial}
          disabled={props.loading || invalid || !props.amount}
        >
          {props.t('Transfer to paid quota')}
        </Button>
        <Button
          variant='outline'
          onClick={props.onAll}
          disabled={props.loading || props.pendingQuota < QUOTA_PER_DOLLAR}
        >
          {props.t('Transfer all')}
        </Button>
      </div>
      <FieldDescription>
        {props.t(
          'The destination is paid quota, not bonus quota. Minimum transfer is one US dollar of quota.'
        )}
      </FieldDescription>
    </Field>
  )
}

function CommissionEmpty(props: { title: string; description: string }) {
  return <EmptyState title={props.title} description={props.description} />
}

function Pager(props: {
  page: number
  total: number
  pageSize: number
  onPageChange: (page: number) => void
  t: (key: string) => string
}) {
  const hasNext = props.page * props.pageSize < props.total
  if (props.page === 1 && !hasNext) return null
  return (
    <div className='mt-3 flex justify-end gap-2'>
      <Button
        variant='outline'
        size='sm'
        disabled={props.page <= 1}
        onClick={() => props.onPageChange(props.page - 1)}
      >
        {props.t('Previous')}
      </Button>
      <Button
        variant='outline'
        size='sm'
        disabled={!hasNext}
        onClick={() => props.onPageChange(props.page + 1)}
      >
        {props.t('Next')}
      </Button>
    </div>
  )
}
