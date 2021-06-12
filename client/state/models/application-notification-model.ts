import { Instance, types } from "mobx-state-tree";

export enum ApplicationNotificationType {
    GIT_CLONE_ERROR = 'git_clone_error'
}

export enum ApplicationNotificationLevel {
    CRITICAL = 'critical'
}

export const ApplicationNotificationModel = types.model({
    uuid       : types.string,
    type       : types.enumeration<ApplicationNotificationType>(Object.values(ApplicationNotificationType)),
    permanent  : types.boolean,
    level      : types.enumeration<ApplicationNotificationLevel>(Object.values(ApplicationNotificationLevel)),
    description: types.string,
    createdAt  : types.string,
});

export interface IApplicationNotification extends Instance<typeof ApplicationNotificationModel> {}