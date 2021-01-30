import { IService } from '@/state/models/service-model';
import { ISession } from '@/state/models/session-model';
import axios from 'axios';
import Axios from 'axios';
import { buildRequest } from './common';

export function retrieveServicesAPI() {
    return buildRequest<IService[]>(() => Axios.get(`/_polo_/api/service`));
}

export function createNewSessionAPI(serviceName: string, checkout: string) {
    return buildRequest<ISession>(() => axios.post(`/_polo_/api/session`, {
        checkout,
        serviceName
    }));
}