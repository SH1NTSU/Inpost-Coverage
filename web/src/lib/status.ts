import type { LockerStatus } from '../api/contracts';

export const STATUS_COLOR: Record<LockerStatus, string> = {
  Operating: '#10b981', 
  Disabled: '#ef4444', 
};

export const STATUS_LABEL: Record<LockerStatus, string> = {
  Operating: 'Operating',
  Disabled: 'Offline',
};
