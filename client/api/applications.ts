import { IApplication } from '@/state/models/application-model';
import { ISession, ISessionLog } from '@/state/models/session-model';
import axios from 'axios';
import Axios from 'axios';
import { buildRequest } from './common';

export function retrieveApplicationsAPI() {
    return buildRequest<IApplication[]>(() => Axios.get(`/_polo_/api/application`));
}

export function createNewSessionAPI(applicationName: string, checkout: string) {
    return buildRequest<ISession>(() => axios.post(`/_polo_/api/session`, {
        checkout,
        applicationName
    }));
}

export function retrieveFailedSessionsAPI() {
    return buildRequest<ISession[]>(() => Axios.get(`/_polo_/api/failed/`));
}

export function retrieveFailedSessionAPI(uuid: string) {
    return buildRequest<ISession[]>(() => Axios.get(`/_polo_/api/failed/${uuid}`));
}

export function retrieveFailedSessionLogsAPI(uuid: string) {
    return buildRequest<ISessionLog[]>(() => Axios.get(`/_polo_/api/failed/${uuid}/logs`));
}