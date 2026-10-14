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
import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { useStatus } from '@/hooks/use-status'
import { useSystemConfig } from '@/hooks/use-system-config'
import { getSelf } from '@/lib/api'
import { handleServerError } from '@/lib/handle-server-error'
import { requireServerSuccess } from '@/lib/server-error-message'

import { getBargainInfo, startBargainActivity } from './api'
import { AffiliateCommissionCard } from './components/affiliate-commission-card'
import { AffiliateRewardsCard } from './components/affiliate-rewards-card'
import { BargainCard } from './components/bargain-card'
import { BillingHistoryDialog } from './components/dialogs/billing-history-dialog'
import { CreemConfirmDialog } from './components/dialogs/creem-confirm-dialog'
import { PaymentConfirmDialog } from './components/dialogs/payment-confirm-dialog'
import { TransferDialog } from './components/dialogs/transfer-dialog'
import { RechargeFormCard } from './components/recharge-form-card'
import { SubscriptionPlansCard } from './components/subscription-plans-card'
import { WalletStatsCard } from './components/wallet-stats-card'
import { DEFAULT_DISCOUNT_RATE, PAYMENT_TYPES } from './constants'
import {
  useTopupInfo,
  usePayment,
  useAffiliate,
  useRedemption,
  useCreemPayment,
  useWaffoPayment,
  useWaffoPancakePayment,
} from './hooks'
import {
  isStripePayment,
  isWaffoPayment,
  isWaffoPancakePayment,
  getDefaultPaymentType,
  getMinTopupAmount,
  dispatchSelectedPayment,
} from './lib'
import type {
  UserWalletData,
  PaymentMethod,
  PresetAmount,
  CreemProduct,
  WaffoPayMethod,
  SpecialTier,
  BargainInfo,
  BargainActivity,
} from './types'

interface WalletProps {
  initialShowHistory?: boolean
}

export function Wallet(props: WalletProps) {
  const { t } = useTranslation()
  const userQuery = useQuery({
    queryKey: ['wallet-user'],
    queryFn: async () =>
      requireServerSuccess(await getSelf()).data as UserWalletData,
  })
  const user = userQuery.data ?? null
  const userLoading = userQuery.isPending
  const [topupAmount, setTopupAmount] = useState(0)
  const [selectedPreset, setSelectedPreset] = useState<number | null>(null)
  const [selectedPaymentMethod, setSelectedPaymentMethod] =
    useState<PaymentMethod>()
  const [selectedWaffoMethodIndex, setSelectedWaffoMethodIndex] = useState<
    number | null
  >(null)
  const [paymentLoading, setPaymentLoading] = useState<string | null>(null)
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false)
  const [transferDialogOpen, setTransferDialogOpen] = useState(false)
  const [billingDialogOpen, setBillingDialogOpen] = useState(
    Boolean(props.initialShowHistory)
  )
  const [redemptionCode, setRedemptionCode] = useState('')
  const [creemDialogOpen, setCreemDialogOpen] = useState(false)
  const [selectedCreemProduct, setSelectedCreemProduct] =
    useState<CreemProduct | null>(null)
  const [showSubscriptionPanel, setShowSubscriptionPanel] = useState(true)
  const [selectedSpecialTier, setSelectedSpecialTier] =
    useState<SpecialTier | null>(null)
  const [selectedBargainActivity, setSelectedBargainActivity] =
    useState<BargainActivity | null>(null)
  const [bargainInfo, setBargainInfo] = useState<BargainInfo | null>(null)
  const [startBargainLoading, setStartBargainLoading] = useState(false)

  const { status } = useStatus()
  const { currency } = useSystemConfig()
  const { topupInfo, presetAmounts, loading: topupLoading } = useTopupInfo()

  // Calculate effective exchange rate - when display type is USD, use rate of 1
  const effectiveUsdExchangeRate = useMemo(() => {
    return currency?.quotaDisplayType === 'USD'
      ? 1
      : currency?.usdExchangeRate || 1
  }, [currency?.quotaDisplayType, currency?.usdExchangeRate])
  const {
    amount: paymentAmount,
    calculating,
    processing,
    calculatePaymentAmount,
    processPayment,
  } = usePayment()
  const {
    affiliateLink,
    loading: affiliateLoading,
    transferQuota,
    transferring,
  } = useAffiliate()
  const { redeeming, redeemCode } = useRedemption()
  const { processing: creemProcessing, processCreemPayment } = useCreemPayment()
  const { processing: waffoProcessing, processWaffoPayment } = useWaffoPayment()
  const { processing: pancakeProcessing, processWaffoPancakePayment } =
    useWaffoPancakePayment()

  const refetchUser = userQuery.refetch
  const fetchUser = useCallback(async () => {
    await refetchUser()
  }, [refetchUser])

  // Fetch bargain info
  const fetchBargainInfo = useCallback(async () => {
    try {
      const res = requireServerSuccess(await getBargainInfo())
      if (res.success && res.data) {
        setBargainInfo(res.data)
      }
    } catch (error) {
      handleServerError(error)
    }
  }, [])

  useEffect(() => {
    if (props.initialShowHistory) {
      window.history.replaceState({}, '', window.location.pathname)
    }
  }, [props.initialShowHistory])

  // Initialize topup amount when topup info is loaded
  const topupAmountInitializedRef = useRef(false)
  useEffect(() => {
    if (topupInfo && !topupAmountInitializedRef.current) {
      topupAmountInitializedRef.current = true
      const minTopup = getMinTopupAmount(topupInfo)
      const defaultPaymentType = getDefaultPaymentType(topupInfo)
      const targetTier = new URLSearchParams(window.location.search).get('tier')
      if (
        topupInfo.special_tiers?.some(
          (tier) => tier.id === targetTier && tier.user_can_purchase
        )
      ) {
        return
      }
      calculatePaymentAmount(minTopup, defaultPaymentType)
        .then(() => setTopupAmount(minTopup))
        .catch(handleServerError)
    }
  }, [topupInfo, calculatePaymentAmount])

  // Get current payment type (selected or default)
  const getCurrentPaymentType = useCallback(() => {
    return selectedPaymentMethod?.type || getDefaultPaymentType(topupInfo)
  }, [selectedPaymentMethod, topupInfo])

  // Handle special tier selection
  const handleSelectSpecialTier = useCallback(
    (tier: SpecialTier | null) => {
      setSelectedBargainActivity(null)
      setSelectedSpecialTier(tier)
      if (tier) {
        setSelectedPreset(null)
        setTopupAmount(tier.price)
        calculatePaymentAmount(tier.price, getCurrentPaymentType(), {
          tier_id: tier.id,
        })
      } else {
        const amount = getMinTopupAmount(topupInfo)
        setTopupAmount(amount)
        calculatePaymentAmount(amount, getCurrentPaymentType())
      }
    },
    [calculatePaymentAmount, getCurrentPaymentType, topupInfo]
  )

  const promoUrlInitialized = useRef(false)

  // Auto select tier from URL search parameter (?tier=xxx)
  useEffect(() => {
    try {
      const params = new URLSearchParams(window.location.search)
      const targetTierId = params.get('tier')
      if (
        !promoUrlInitialized.current &&
        targetTierId &&
        Array.isArray(topupInfo?.special_tiers)
      ) {
        promoUrlInitialized.current = true
        const matchedTier = topupInfo.special_tiers.find(
          (t) => t.id === targetTierId && t.user_can_purchase
        )
        if (matchedTier) {
          calculatePaymentAmount(matchedTier.price, getCurrentPaymentType(), {
            tier_id: matchedTier.id,
          })
            .then(() => handleSelectSpecialTier(matchedTier))
            .catch(handleServerError)
        }
      }
    } catch {
      // ignore
    }
  }, [
    topupInfo?.special_tiers,
    handleSelectSpecialTier,
    calculatePaymentAmount,
    getCurrentPaymentType,
  ])

  // Handle start bargain
  const handleStartBargainActivity = useCallback(async () => {
    setStartBargainLoading(true)
    try {
      const res = requireServerSuccess(await startBargainActivity())
      if (res.success) {
        toast.success(
          t('Bargain activity started! Invite friends to slash the price')
        )
        await fetchBargainInfo()
      } else {
        toast.error(res.message || t('Failed to start bargain activity'))
      }
    } catch (error) {
      handleServerError(error)
    } finally {
      setStartBargainLoading(false)
    }
  }, [fetchBargainInfo, t])

  // Handle buy bargain
  const handleBuyBargain = useCallback(
    (activity: BargainActivity) => {
      if (!activity) return
      setSelectedSpecialTier(null)
      setSelectedPreset(null)
      setSelectedBargainActivity(activity)
      setTopupAmount(activity.current_price)
      calculatePaymentAmount(activity.current_price, getCurrentPaymentType(), {
        bargain_activity_id: activity.id,
      })
      document
        .querySelector('#wallet-add-funds')
        ?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    },
    [calculatePaymentAmount, getCurrentPaymentType]
  )

  // Handle preset selection
  const handleSelectPreset = (preset: PresetAmount) => {
    setSelectedSpecialTier(null)
    setSelectedBargainActivity(null)
    setTopupAmount(preset.value)
    setSelectedPreset(preset.value)
    calculatePaymentAmount(preset.value, getCurrentPaymentType())
  }

  // Handle topup amount change
  const handleTopupAmountChange = (amount: number) => {
    setSelectedSpecialTier(null)
    setSelectedBargainActivity(null)
    setTopupAmount(amount)
    setSelectedPreset(null)
    calculatePaymentAmount(amount, getCurrentPaymentType())
  }

  // Handle payment method selection
  const handlePaymentMethodSelect = async (method: PaymentMethod) => {
    if (
      (selectedSpecialTier || selectedBargainActivity) &&
      (isStripePayment(method.type) ||
        isWaffoPayment(method.type) ||
        isWaffoPancakePayment(method.type))
    ) {
      toast.error(t('Promotional packs only support online payment'))
      return
    }
    setSelectedPaymentMethod(method)
    setSelectedWaffoMethodIndex(null)
    setPaymentLoading(method.type)

    try {
      // Validate minimum topup only for standard (non-special-tier and non-bargain) topup
      if (!selectedSpecialTier && !selectedBargainActivity) {
        const minTopup = getMinTopupAmount(topupInfo)
        if (topupAmount < minTopup) {
          return
        }
      }

      // Calculate payment amount and show confirmation dialog
      await calculatePaymentAmount(topupAmount, method.type, {
        tier_id: selectedSpecialTier?.id,
        bargain_activity_id: selectedBargainActivity?.id,
      })
      setConfirmDialogOpen(true)
    } finally {
      setPaymentLoading(null)
    }
  }

  // Handle payment confirmation
  const handlePaymentConfirm = async () => {
    if (!selectedPaymentMethod) return

    const options = {
      tier_id: selectedSpecialTier ? selectedSpecialTier.id : undefined,
      bargain_activity_id: selectedBargainActivity
        ? selectedBargainActivity.id
        : undefined,
    }

    const success = await dispatchSelectedPayment(
      selectedPaymentMethod,
      topupAmount,
      selectedWaffoMethodIndex,
      {
        regular: (amt, type) => processPayment(amt, type, options),
        waffo: processWaffoPayment,
        waffoPancake: processWaffoPancakePayment,
      },
      options
    )

    if (success) {
      setConfirmDialogOpen(false)
      setSelectedSpecialTier(null)
      setSelectedBargainActivity(null)
      await fetchUser()
      if (options.bargain_activity_id) {
        await fetchBargainInfo()
      }
    }
  }

  // Handle redemption
  const handleRedeem = async () => {
    if (!redemptionCode) return

    const success = await redeemCode(redemptionCode)
    if (success) {
      setRedemptionCode('')
      await fetchUser()
    }
  }

  // Handle transfer
  const handleTransfer = async (amount: number) => {
    const success = await transferQuota(amount)
    if (success) {
      await fetchUser()
    }
    return success
  }

  // Handle Creem product selection
  const handleCreemProductSelect = (product: CreemProduct) => {
    setSelectedCreemProduct(product)
    setCreemDialogOpen(true)
  }

  // Handle Creem payment confirmation
  const handleCreemConfirm = async () => {
    if (!selectedCreemProduct) return

    const success = await processCreemPayment(selectedCreemProduct.productId)
    if (success) {
      setCreemDialogOpen(false)
      setSelectedCreemProduct(null)
      await fetchUser()
    }
  }

  const handleWaffoMethodSelect = async (
    method: WaffoPayMethod,
    index: number
  ) => {
    if (selectedSpecialTier || selectedBargainActivity) return
    const loadingKey = `waffo-${index}`
    setSelectedPaymentMethod({
      name: method.name,
      type: PAYMENT_TYPES.WAFFO,
      icon: method.icon,
    })
    setSelectedWaffoMethodIndex(index)
    setPaymentLoading(loadingKey)

    try {
      await calculatePaymentAmount(topupAmount, PAYMENT_TYPES.WAFFO)
      setConfirmDialogOpen(true)
    } finally {
      setPaymentLoading(null)
    }
  }

  // Get discount rate for current topup amount
  const getDiscountRate = useCallback(() => {
    return topupInfo?.discount?.[topupAmount] || DEFAULT_DISCOUNT_RATE
  }, [topupInfo, topupAmount])

  const handleSubscriptionAvailabilityChange = useCallback(
    (available: boolean) => {
      setShowSubscriptionPanel(available)
    },
    []
  )

  const itemDescription = useMemo(() => {
    if (selectedSpecialTier) {
      return `${selectedSpecialTier.name} (¥${selectedSpecialTier.price})`
    }
    if (selectedBargainActivity) {
      return `${selectedBargainActivity.title} (${t('Slashed to')} ¥${selectedBargainActivity.current_price.toFixed(2)})`
    }
    return undefined
  }, [selectedSpecialTier, selectedBargainActivity, t])

  return (
    <>
      <SectionPageLayout>
        <SectionPageLayout.Title>{t('Wallet')}</SectionPageLayout.Title>
        <SectionPageLayout.Content>
          <div className='mx-auto flex w-full max-w-7xl flex-col gap-4 sm:gap-5'>
            <WalletStatsCard user={user} loading={userLoading} />

            <div
              className={
                showSubscriptionPanel
                  ? 'grid gap-4 xl:grid-cols-[minmax(0,1.05fr)_minmax(360px,0.95fr)] xl:items-start'
                  : 'grid gap-4'
              }
            >
              <div
                id='wallet-add-funds'
                className='flex scroll-mt-4 flex-col gap-4'
              >
                <BargainCard
                  bargainInfo={bargainInfo || topupInfo?.bargain}
                  onStartActivity={handleStartBargainActivity}
                  onBuyBargain={handleBuyBargain}
                  onRefresh={fetchBargainInfo}
                  affLink={affiliateLink}
                  startLoading={startBargainLoading}
                />
                <RechargeFormCard
                  topupInfo={topupInfo}
                  presetAmounts={presetAmounts}
                  specialTiers={topupInfo?.special_tiers}
                  selectedSpecialTier={selectedSpecialTier}
                  promotionalPurchase={Boolean(
                    selectedSpecialTier || selectedBargainActivity
                  )}
                  onSelectSpecialTier={handleSelectSpecialTier}
                  selectedPreset={selectedPreset}
                  onSelectPreset={handleSelectPreset}
                  topupAmount={topupAmount}
                  onTopupAmountChange={handleTopupAmountChange}
                  paymentAmount={paymentAmount}
                  calculating={calculating}
                  onPaymentMethodSelect={handlePaymentMethodSelect}
                  paymentLoading={paymentLoading}
                  redemptionCode={redemptionCode}
                  onRedemptionCodeChange={setRedemptionCode}
                  onRedeem={handleRedeem}
                  redeeming={redeeming}
                  topupLink={topupInfo?.topup_link}
                  loading={topupLoading}
                  priceRatio={(status?.price as number) || 1}
                  usdExchangeRate={effectiveUsdExchangeRate}
                  onOpenBilling={() => setBillingDialogOpen(true)}
                  creemProducts={topupInfo?.creem_products}
                  enableCreemTopup={topupInfo?.enable_creem_topup}
                  onCreemProductSelect={handleCreemProductSelect}
                  enableWaffoTopup={topupInfo?.enable_waffo_topup}
                  waffoPayMethods={topupInfo?.waffo_pay_methods}
                  waffoMinTopup={topupInfo?.waffo_min_topup}
                  onWaffoMethodSelect={handleWaffoMethodSelect}
                  enableWaffoPancakeTopup={
                    topupInfo?.enable_waffo_pancake_topup
                  }
                />
              </div>

              <SubscriptionPlansCard
                topupInfo={topupInfo}
                onAvailabilityChange={handleSubscriptionAvailabilityChange}
                userQuota={user?.quota}
                onPurchaseSuccess={fetchUser}
              />
            </div>

            <AffiliateRewardsCard
              user={user}
              affiliateLink={affiliateLink}
              onTransfer={() => setTransferDialogOpen(true)}
              complianceConfirmed={
                topupInfo?.payment_compliance_confirmed !== false
              }
              loading={affiliateLoading}
            />

            <AffiliateCommissionCard onPaidQuotaChanged={fetchUser} />
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <PaymentConfirmDialog
        open={confirmDialogOpen}
        onOpenChange={setConfirmDialogOpen}
        onConfirm={handlePaymentConfirm}
        topupAmount={topupAmount}
        paymentAmount={paymentAmount}
        paymentMethod={selectedPaymentMethod}
        calculating={calculating}
        processing={processing || waffoProcessing || pancakeProcessing}
        discountRate={getDiscountRate()}
        usdExchangeRate={effectiveUsdExchangeRate}
        itemDescription={itemDescription}
      />

      <TransferDialog
        open={transferDialogOpen}
        onOpenChange={setTransferDialogOpen}
        onConfirm={handleTransfer}
        availableQuota={user?.aff_quota ?? 0}
        transferring={transferring}
      />

      <BillingHistoryDialog
        open={billingDialogOpen}
        onOpenChange={setBillingDialogOpen}
      />

      <CreemConfirmDialog
        open={creemDialogOpen}
        onOpenChange={setCreemDialogOpen}
        onConfirm={handleCreemConfirm}
        product={selectedCreemProduct}
        processing={creemProcessing}
      />
    </>
  )
}
