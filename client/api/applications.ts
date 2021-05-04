import { IApplication } from '@/state/models/application-model';
import { ISession } from '@/state/models/session-model';
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