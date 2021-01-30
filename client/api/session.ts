import { ISession } from '@/state/models/session-model';
import Axios from 'axios';
import { buildRequest } from './common';

export function retrieveAllSessionsAPI() {
    return buildRequest<ISession[]>(() => Axios.get(`/_polo_/api/session/`));
}

export function retrieveSessionAPI(uuid: string) {
    return buildRequest<ISession>(() => Axios.get(`/_polo_/api/session/${uuid}`));
}

export function trackSessionAPI(uuid: string) {
    return buildRequest<void>(() => Axios.post(`/_polo_/api/session/${uuid}/track`));
}