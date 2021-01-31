import { ISession, ISessionLog } from '@/state/models/session-model';
import Axios from 'axios';
import { buildRequest } from './common';

export interface IAPISession extends Omit<ISession, 'logs'> {
    logs: ISessionLog[];
}

export function retrieveAllSessionsAPI() {
    return buildRequest<IAPISession[]>(() => Axios.get(`/_polo_/api/session/`));
}

export function killSessionAPI(uuid: string) {
    return buildRequest<void>(() => Axios.delete(`/_polo_/api/session/${uuid}`));
}

export function retrieveSessionAPI(uuid: string) {
    return buildRequest<IAPISession>(() => Axios.get(`/_polo_/api/session/${uuid}`));
}

export function trackSessionAPI(uuid: string) {
    return buildRequest<void>(() => Axios.post(`/_polo_/api/session/${uuid}/track`));
}

export function untrackSessionAPI() {
    return buildRequest<void>(() => Axios.delete(`/_polo_/api/session/<none>/track`))
}

export function retrieveSessionAgeAPI(uuid: string) {
    return buildRequest<number>(() => Axios.get(`/_polo_/api/session/${uuid}/age`));
}