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
const HUB_URL = 'https://gift.inkwan.app'

interface ActivityFeature {
  icon: string
  text: string
}

interface ActivityItem {
  id: string
  href: string
  theme: string
  tabName: string
  tabIcon: string
  badge: string
  kicker: string
  title: string
  subtitle: string
  features: ActivityFeature[]
  ctaText: string
  bannerType: 'image' | 'picture' | 'composite-models'
  bannerSrc: string
  bannerWebp?: string
  bannerFallback: string
}

export const HOME_ACTIVITIES: ActivityItem[] = [
  {
    id: 'lottery',
    href: `${HUB_URL}/lottery`,
    theme: 'theme-lottery',
    tabName: 'Lucky draw',
    tabIcon: '🎁',
    badge: 'Instant draw on recharge',
    kicker: 'LUCKY DRAW · ROULETTE',
    title: 'Recharge lucky draw',
    subtitle:
      'Hanako-kun supernatural theme! Single recharge of ¥10+ earns 1 free lucky wheel draw with full refunds and real quotas.',
    features: [
      { icon: '🎯', text: 'Single qualifying recharge earns 1 draw chance' },
      {
        icon: '👑',
        text: 'Chance to win full-refund equivalent codes and cash quotas',
      },
      {
        icon: '⚡',
        text: 'Direct integration with waner~ API站 to auto-generate redemption keys',
      },
    ],
    ctaText: 'Enter draw now',
    bannerType: 'image',
    bannerSrc: '/img/activities/lottery_hero.png',
    bannerFallback: 'https://gift.inkwan.app/img/lottery_hero.png',
  },
  {
    id: 'invite',
    href: `${HUB_URL}/invite`,
    theme: 'theme-invite',
    tabName: 'Invite',
    tabIcon: '👤',
    badge: '10% Lifetime rebate',
    kicker: 'AFFILIATE · PROGRAM',
    title: 'Developer Affiliate Rebate',
    subtitle:
      'Share your referral link with friends and get a 10% quota rebate in seconds upon their recharge. Permanent lock-in, universal across models!',
    features: [
      {
        icon: '💎',
        text: '10% recharge amount credited instantly as real quota',
      },
      {
        icon: '🔗',
        text: 'One-click account binding to generate your exclusive referral link',
      },
      {
        icon: '🚀',
        text: 'Site-wide model compute deduction with no threshold',
      },
    ],
    ctaText: 'Get referral link',
    bannerType: 'picture',
    bannerWebp: '/img/activities/invite/hero_invite.webp',
    bannerSrc: '/img/activities/invite/hero_invite.png',
    bannerFallback: 'https://gift.inkwan.app/img/invite/hero_invite.png',
  },
  {
    id: 'tiers',
    href: `${HUB_URL}/tiers`,
    theme: 'theme-tiers',
    tabName: 'Special Tiers',
    tabIcon: '⚡',
    badge: 'Recharge 5 get 5 bonus',
    kicker: 'SPECIAL TIERS · BONUS',
    title: 'Recharge Special Bonus Tiers',
    subtitle:
      'Limited-time compute carnival! Recharge 5 get 5 with 100% bonus, recharge 100 get 15, recharge 200 get 40. Universal across 40+ models!',
    features: [
      {
        icon: '⚡',
        text: 'Recharge 5 get 5 with 100% instant quota bonus, 1 per user',
      },
      {
        icon: '💎',
        text: 'Large amount tier saves ¥40 directly with multi-buy support',
      },
      {
        icon: '🚀',
        text: 'Second-level auto credit, site-wide 40+ premier models universal',
      },
    ],
    ctaText: 'View special tiers',
    bannerType: 'image',
    bannerSrc: '/img/activities/tiers_hero.jpg',
    bannerFallback: 'https://gift.inkwan.app/img/tiers_hero.jpg',
  },
  {
    id: 'models',
    href: `${HUB_URL}/models`,
    theme: 'theme-models',
    tabName: 'Model Square',
    tabIcon: '📖',
    badge: '40+ Top models',
    kicker: 'CATALOGUE · PRICING',
    title: 'Site-wide Model Pricing',
    subtitle:
      'Transparent multiplier matrix! 40+ premier models from OpenAI, Anthropic, Google, DeepSeek with dedicated free model zone.',
    features: [
      {
        icon: '⭐',
        text: 'Transparent pay-per-call and pay-per-token pricing rates',
      },
      {
        icon: '🎁',
        text: 'Quick lookup for free gift models and trending models',
      },
      {
        icon: '📋',
        text: 'One-click copy model code for instant project integration',
      },
    ],
    ctaText: 'Browse model prices',
    bannerType: 'composite-models',
    bannerSrc: '/img/activities/models/title_banner.png',
    bannerFallback: 'https://gift.inkwan.app/img/models/title_banner.png',
  },
]
