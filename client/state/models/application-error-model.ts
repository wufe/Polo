import { Instance, types } from "mobx-state-tree";

export enum ApplicationErrorType {
    GIT_CREDENTIALS_ERROR = 'git_credentials_error'
}

export const ApplicationErrorModel = types.model({
    uuid       : types.string,
    type       : types.enumeration<ApplicationErrorType>(Object.values(ApplicationErrorType)),
    description: types.string,
    createdAt  : types.string,
});

export interface IApplicationError extends Instance<typeof ApplicationErrorModel> {}