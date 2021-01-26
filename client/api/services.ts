import { IService } from '@/state/models/service';
import Axios from 'axios';
import { buildRequest } from './common';

export function retrieveServicesAPI() {
    return buildRequest<IService[]>(() => Axios.get(`/_polo_/api/service`));
}