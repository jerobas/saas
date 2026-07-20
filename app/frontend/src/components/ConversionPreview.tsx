import type { MeasurementUnitResponse } from "../gateways/desktopBridge";

interface ConversionPreviewProps {
  label: string;
  numeratorAtomic: number | string;
  denominator: number | string;
  baseUnit: MeasurementUnitResponse;
  className?: string;
}

const integerFormatter = new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 0 });
const decimalFormatter = new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 6 });

const greatestCommonDivisor = (left: number, right: number) => {
  let a = Math.abs(left);
  let b = Math.abs(right);
  while (b !== 0) [a, b] = [b, a % b];
  return a;
};

const hasFiniteDecimal = (denominator: number) => {
  let remaining = denominator;
  while (remaining % 2 === 0) remaining /= 2;
  while (remaining % 5 === 0) remaining /= 5;
  return remaining === 1;
};

const formatBaseUnitEquivalent = (
  numeratorAtomic: number | string,
  denominator: number | string,
  baseUnit: MeasurementUnitResponse,
) => {
  const numeratorValue = Number(numeratorAtomic);
  const denominatorValue = Number(denominator);
  const equivalentNumerator = numeratorValue * baseUnit.denominator;
  const equivalentDenominator = denominatorValue * baseUnit.numeratorAtomic;

  if (
    !Number.isSafeInteger(equivalentNumerator) ||
    !Number.isSafeInteger(equivalentDenominator) ||
    equivalentNumerator <= 0 ||
    equivalentDenominator <= 0
  ) {
    return null;
  }

  const divisor = greatestCommonDivisor(equivalentNumerator, equivalentDenominator);
  const reducedNumerator = equivalentNumerator / divisor;
  const reducedDenominator = equivalentDenominator / divisor;
  const quantity =
    reducedDenominator === 1
      ? integerFormatter.format(reducedNumerator)
      : hasFiniteDecimal(reducedDenominator)
        ? decimalFormatter.format(reducedNumerator / reducedDenominator)
        : `${integerFormatter.format(reducedNumerator)}/${integerFormatter.format(reducedDenominator)}`;

  return `${quantity} ${baseUnit.symbol}`;
};

function ConversionPreview({
  label,
  numeratorAtomic,
  denominator,
  baseUnit,
  className = "text-xs text-slate-500",
}: ConversionPreviewProps) {
  const equivalent = formatBaseUnitEquivalent(numeratorAtomic, denominator, baseUnit);
  if (!equivalent) return null;

  return (
    <p className={className}>
      {label} = <strong>{equivalent}</strong>
    </p>
  );
}

export default ConversionPreview;
