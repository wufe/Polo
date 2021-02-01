import { APIPayload } from "@/api/common";
import { createNewSessionAPI } from "@/api/services";
import { flow, Instance, SnapshotIn, SnapshotOut, types } from "mobx-state-tree";
import { ISession } from "./session-model";

export const ServiceModel = types.model({
    name                 : types.string,
    remote               : types.string,
    target               : types.string,
    host                 : types.string,
    maxConcurrentSessions: types.number,
    serviceFolder        : types.string,
    branches             : types.array(types.string)
})
.actions(self => {

    const newSession = flow(function* newSession(checkout: string) {
        const session: APIPayload<ISession> = yield  createNewSessionAPI(self.name, checkout);
        return session;
    });

    return { newSession };
})

export interface IService extends Instance<typeof ServiceModel> { }
export interface IServiceSnapshotOut extends SnapshotOut<typeof ServiceModel> { }
export interface IServiceSnapshotIn extends SnapshotIn<typeof ServiceModel> { }