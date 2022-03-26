import { useEffect, useRef, useState } from "react";
import { has, observe, values } from "mobx";
import { IApplication, TNotificationResult } from "@polo/common/state/models";
import { ApplicationNotificationLevel, IApplicationNotification } from "@polo/common/state/models/application-notification-model";
import { useNotification } from "@polo/common/state/models/notification-hook";
import { NotificationType } from "@polo/common/state/models/notification-model";
import { TDictionary } from "@polo/common/utils/types";

export type TApplicationNotification = TNotificationResult & {
    appNotification: IApplicationNotification;
}

export const useApplicationNotifications = (application: IApplication) => {

    type TApplicationNotificationsMap = TDictionary<IApplicationNotification>;
    const [notifications, setNotifications] = useState<IApplicationNotification[]>(() => values(application.notifications) as unknown as IApplicationNotification[]);
    const { notify } = useNotification();
    const showedNotifications = useRef<TDictionary<TApplicationNotification>>({});

    useEffect(() => {
        setNotifications(values(application.notifications) as unknown as IApplicationNotification[]);
        const disposer = observe(application.notifications, ({ object }) => {
            setNotifications(values(object) as IApplicationNotification[]);
        });
        return disposer;
    }, [application]);

    useEffect(() => {
        const uniqueNotificationsByType = (notifications as IApplicationNotification[])
            .reduce<TApplicationNotificationsMap>((acc, notification) => {
                if (!acc[notification.type]) {
                    acc[notification.type] = notification;
                }
                return acc;
            }, {});
        const uniqueNotificationByID = (notifications as IApplicationNotification[])
            .reduce<TApplicationNotificationsMap>((acc, notification) => {
                if (!acc[notification.uuid]) {
                    acc[notification.uuid] = notification;
                }
                return acc;
            }, {});
        const uniqueNotifications = Object.values(uniqueNotificationsByType);

        Object.entries(showedNotifications.current)
            .forEach(([_, v]) => {
                const { appNotification, remove } = v;
                if (!uniqueNotificationByID[appNotification.uuid]) {
                    remove();
                    delete showedNotifications.current[appNotification.uuid];
                }
            });
        uniqueNotifications
            .forEach(notification => {
                if (!showedNotifications.current[notification.uuid]) {
                    let type: NotificationType = NotificationType.INFO;
                    switch (notification.level) {
                        case ApplicationNotificationLevel.CRITICAL:
                            type = NotificationType.ERROR;
                            break;
                    }
                    let expiration = notification.permanent ? 0 : 10;
                    const appNotification: TApplicationNotification = {
                        ...notify({
                            text: notification.description,
                            type,
                            expiration
                        }),
                        appNotification: notification
                    };
                    showedNotifications.current[notification.uuid] = appNotification;
                }
            });
    }, [notifications]);
}