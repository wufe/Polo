import Axios from 'axios';
import { IService } from "@/state/models";

export const retrieveServicesAPI = async (): Promise<IService[]> => {
    const response = await Axios.get(`/_polo_/api/service`);
    return response.data;
}