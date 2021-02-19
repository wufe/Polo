import { createContext } from 'react';

export enum HelperStatus {
    RUNNING = 'running',
    EXPIRED = 'expired',
}

export const HelperStatusContext = createContext<{ status: HelperStatus; age: number }>({ status: HelperStatus.RUNNING, age: 0 });