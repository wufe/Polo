import { ISession } from '@/state/models/session';
import Axios from 'axios';
import { buildRequest } from './common';

export function retrieveSessionAPI(uuid: string) {
    return buildRequest<ISession>(() => Axios.get(`/_polo_/api/session/${uuid}`));
}