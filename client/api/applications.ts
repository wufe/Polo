import { IApplication } from '@/state/models/application-model';
import { ISession, ISessionLog } from '@/state/models/session-model';
import axios from 'axios';
import Axios from 'axios';
import { buildRequest } from './common';
import { IAPISession } from './session';

export interface IAPIFailedSessions {
    acknowledged: ISession[];
    unacknowledged: ISession[];
}

export interface IAPIStatusData {
    applications: IApplication[];
    sessions: IAPISession[];
    failures: IAPIFailedSessions;
}

export function retrieveStatusDataAPI() {
    return buildRequest<IAPIStatusData>(() => Axios.get(`/_polo_/api/status`));
}

export function createNewSessionAPI(applicationName: string, checkout: string) {
    return buildRequest<ISession>(() => axios.post(`/_polo_/api/session`, {
        checkout,
        applicationName
    }));
}

export function retrieveFailedSessionAPI(uuid: string) {
    return buildRequest<ISession[]>(() => Axios.get(`/_polo_/api/failed/${uuid}`));
}

export function markFailedSessionAsAcknowledgedAPI(uuid: string) {
    return buildRequest<void>(() => Axios.post(`/_polo_/api/failed/${uuid}/ack`));
}

export function retrieveFailedSessionLogsAPI(uuid: string) {
    return buildRequest<ISessionLog[]>(() => Axios.get(`/_polo_/api/failed/${uuid}/logs`));
}