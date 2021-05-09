import { getParent, hasParent, Instance, types } from "mobx-state-tree";
import { IApp } from "./app-model";

export enum NotificationType {
    SUCCESS = 'success',
    ERROR   = 'error',
    INFO    = 'info',
}

export const NotificationModel = types.model({
    expiration: types.number,
    text      : types.string,
    title     : types.optional(types.string, ''),
    type      : types.enumeration<NotificationType>(Object.values(NotificationType)),
    uuid      : types.string,
})
.actions(self => {
    
    const remove = () => {
        if (hasParent(self, 2)) {
            (getParent(self, 2) as IApp).deleteNotification(self.uuid);
        }
    };

    let onClick: (notification: INotification) => void;
    const addOnClick = (cb: typeof onClick) => {
        onClick = cb;
    };

    const triggerOnClick = () => {
        if (onClick) {
            onClick(self as INotification);
        }
    };

    return {
        addOnClick,
        triggerOnClick,
        remove
    };
});

export interface INotification extends Instance<typeof NotificationModel> {}