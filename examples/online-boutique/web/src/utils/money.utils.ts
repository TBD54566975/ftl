import { Money } from '../api/currency'

export const formatMoney = (money: Money): string => {
  const fractionalAmount = money.nanos / 1e9
  const totalAmount = money.units + fractionalAmount
  return `$${totalAmount.toFixed(2)}`
}
