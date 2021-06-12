import { APIPayload, APIRequestResult } from "@/api/common";
import { createNewSessionAPI, IAPIApplication } from "@/api/applications";
import { flow, Instance, SnapshotIn, SnapshotOut, types } from "mobx-state-tree";
import { ISession, SessionModel } from "./session-model";
import { ApplicationErrorModel, IApplicationError } from "./application-error-model";
import { TDictionary } from "@/utils/types";

const checkoutObject = {
    name       : types.string,
    hash       : types.string,
    author     : types.string,
    authorEmail: types.string,
    date       : types.string,
    message    : types.string,
};

export const ApplicationBranchModel = types.model({
    ...checkoutObject
});

export const ApplicationTagModel = types.model({
    ...checkoutObject
});

export interface IApplicationBranchModel extends Instance<typeof ApplicationBranchModel> {}

export const ApplicationConfigurationModel = types.model({
    id                   : types.string,
    name                 : types.string,
    hash                 : types.string,
    remote               : types.string,
    target               : types.string,
    host                 : types.string,
    maxConcurrentSessions: types.number,
})

export const ApplicationModel = types.model({
    filename      : types.string,
    configuration : ApplicationConfigurationModel,
    folder        : types.string,
    branchesMap   : types.map(ApplicationBranchModel),
    tagsMap       : types.map(ApplicationTagModel),
    failedSessions: types.map(SessionModel),
    errors        : types.map(ApplicationErrorModel),
})
.actions(self => {

    const newSession = flow(function* newSession(checkout: string) {
        const session: APIPayload<ISession> = yield createNewSessionAPI(self.configuration.name, checkout);
        return session;
    });

    return { newSession };
})

export interface IApplication extends Instance<typeof ApplicationModel> { }
export interface IApplicationSnapshotOut extends SnapshotOut<typeof ApplicationModel> { }
export interface IApplicationSnapshotIn extends SnapshotIn<typeof ApplicationModel> { }

export const castAPIApplicationToApplicationModel = (apiApplication: IAPIApplication): IApplication => {
    const { errors, ...rest } = apiApplication;
    const application = rest as IApplication;
    if (errors && errors.length) {
        type TApplicationErrorsMap = TDictionary<IApplicationError>;
        application.errors = errors.reduce<TApplicationErrorsMap>((acc, error) => {
            acc[error.uuid] = error;
            return acc;
        }, {}) as any;
    }
    return application;
}