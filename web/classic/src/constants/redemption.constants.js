/*
Copyright (C) 2025 QuantumNous

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

export const REDEMPTION_STATUS = {
  UNUSED: 1, // Unused
  DISABLED: 2, // Disabled
  USED: 3, // Used
};

export const REDEMPTION_QUOTA_TYPE = {
  PAID: 'paid',
  BONUS: 'bonus',
};

export const REDEMPTION_QUOTA_TYPES = {
  [REDEMPTION_QUOTA_TYPE.PAID]: {
    color: 'grey',
    text: '付费额度',
    optionText: '增加额度',
  },
  [REDEMPTION_QUOTA_TYPE.BONUS]: {
    color: 'green',
    text: '免费积分',
    optionText: '增加免费积分',
  },
};

export const normalizeRedemptionQuotaType = (quotaType) => {
  return quotaType === REDEMPTION_QUOTA_TYPE.BONUS
    ? REDEMPTION_QUOTA_TYPE.BONUS
    : REDEMPTION_QUOTA_TYPE.PAID;
};

// Redemption code status display mapping
export const REDEMPTION_STATUS_MAP = {
  [REDEMPTION_STATUS.UNUSED]: {
    color: 'green',
    text: '未使用',
  },
  [REDEMPTION_STATUS.DISABLED]: {
    color: 'red',
    text: '已禁用',
  },
  [REDEMPTION_STATUS.USED]: {
    color: 'grey',
    text: '已使用',
  },
};

// Action type constants
export const REDEMPTION_ACTIONS = {
  DELETE: 'delete',
  ENABLE: 'enable',
  DISABLE: 'disable',
};
