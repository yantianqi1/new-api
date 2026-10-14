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
import { act, fireEvent, render, screen, within } from '@testing-library/react'
import { createInstance, type i18n } from 'i18next'
import { I18nextProvider, initReactI18next } from 'react-i18next'
import { beforeEach, expect, test } from 'vitest'

import en from '@/i18n/locales/en.json'
import fr from '@/i18n/locales/fr.json'
import ja from '@/i18n/locales/ja.json'
import ru from '@/i18n/locales/ru.json'
import vi from '@/i18n/locales/vi.json'
import zhTW from '@/i18n/locales/zh-TW.json'
import zhCN from '@/i18n/locales/zh.json'

import { HOME_ACTIVITIES } from '../components/waner/activity-data'
import ActivityShowcase from '../components/waner/activity-showcase'

let language: i18n
beforeEach(async () => {
  language = createInstance()
  await language.use(initReactI18next).init({
    lng: 'zhCN',
    fallbackLng: 'en',
    resources: { en, zhCN, zhTW, fr, ru, ja, vi },
    interpolation: { escapeValue: false },
  })
})

test('Chinese activity navigation and all four offers render localized copy', () => {
  render(
    <I18nextProvider i18n={language}>
      <ActivityShowcase />
    </I18nextProvider>
  )
  for (const offer of [
    { tab: '充值抽奖', title: '充值幸运转盘', action: '立即抽奖' },
    { tab: '邀请返佣', title: '邀请好友，赚取返佣', action: '获取邀请链接' },
    { tab: '充值优惠', title: '限时充值加赠', action: '查看充值档位' },
    { tab: '模型广场', title: '全站模型价格', action: '浏览模型价格' },
  ]) {
    const tab = screen.getByRole('tab', { name: new RegExp(offer.tab) })
    fireEvent.click(tab)
    expect(tab).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByRole('heading', { name: offer.title })).toBeVisible()
    expect(screen.getByRole('link', { name: offer.action })).toBeVisible()
    expect(screen.getByRole('img', { name: offer.title })).toBeVisible()
  }
})

test('switching language keeps the selected activity and updates its copy', async () => {
  render(
    <I18nextProvider i18n={language}>
      <ActivityShowcase />
    </I18nextProvider>
  )
  fireEvent.click(screen.getByRole('tab', { name: /邀请返佣/ }))
  await act(() => language.changeLanguage('en'))
  expect(screen.getByRole('tab', { name: /Invite/ })).toHaveAttribute(
    'aria-selected',
    'true'
  )
  expect(
    screen.getByRole('heading', { name: 'Developer Affiliate Rebate' })
  ).toBeVisible()
  await act(() => language.changeLanguage('zhTW'))
  expect(screen.getByRole('tab', { name: /邀請返佣/ })).toHaveAttribute(
    'aria-selected',
    'true'
  )
  expect(
    screen.getByRole('heading', { name: '邀請好友，賺取返佣' })
  ).toBeVisible()
})

test('selected activity has an accessible panel with its offer and navigation', () => {
  render(
    <I18nextProvider i18n={language}>
      <ActivityShowcase />
    </I18nextProvider>
  )
  const panel = screen.getByRole('tabpanel')
  expect(within(panel).getByRole('link', { name: '立即抽奖' })).toHaveAttribute(
    'href',
    'https://gift.inkwan.app/lottery'
  )
  fireEvent.click(within(panel).getByRole('button', { name: '下一个活动' }))
  expect(screen.getByRole('tab', { name: /邀请返佣/ })).toHaveAttribute(
    'aria-selected',
    'true'
  )
})

test.each(['zhCN', 'zhTW', 'fr', 'ru', 'ja', 'vi'])(
  '%s renders translated badges, descriptions, features and actions for every offer',
  async (locale) => {
    await language.changeLanguage(locale)
    render(
      <I18nextProvider i18n={language}>
        <ActivityShowcase />
      </I18nextProvider>
    )
    for (const offer of HOME_ACTIVITIES) {
      fireEvent.click(
        screen.getByRole('tab', {
          name: language.t(offer.tabName),
        })
      )
      const panel = screen.getByRole('tabpanel')
      for (const source of [
        offer.badge,
        offer.kicker,
        offer.title,
        offer.subtitle,
        offer.ctaText,
        ...offer.features.map((feature) => feature.text),
      ]) {
        const translated = language.getResource(locale, 'translation', source)
        expect(translated).toBeTypeOf('string')
        expect(translated).not.toBe(source)
        expect(panel).toHaveTextContent(translated)
      }
    }
  }
)
