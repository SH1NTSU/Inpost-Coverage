import type { ReactNode } from 'react';

interface Props {
  label: string;
  value: ReactNode;
  hint?: string;
  tone?: 'neutral' | 'good' | 'bad';
}

export function StatPill({ label, value, hint, tone = 'neutral' }: Props) {
  return (
    <div className={`stat-pill stat-pill--${tone}`}>
      <div className="stat-pill__label">{label}</div>
      <div className="stat-pill__value">{value}</div>
      {hint && <div className="stat-pill__hint">{hint}</div>}
    </div>
  );
}
