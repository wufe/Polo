import { SessionStatus, SessionKillReason } from '@polo/common/state/models/session-model-enums';
import { createContext } from 'react';

export enum HelperStatus {
    RUNNING  = 'running',
    EXPIRED  = 'expired',
    REPLACED = 'replaced',
}

export const HelperStatusContext = createContext<{
    helperStatus: HelperStatus;
    age: number;
    status: SessionStatus;
    killReason: SessionKillReason;
    replacedBy: string;
}>({ helperStatus: HelperStatus.RUNNING, age: 0, status: SessionStatus.NONE, replacedBy: '', killReason: SessionKillReason.NONE });