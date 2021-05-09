import React from 'react';
import { Notification } from './notification';
import { INotification, NotificationType } from '@/state/models/notification-model';

type TProps = {
    notifications: INotification[];
};
export const NotificationsContainer = (props: React.PropsWithChildren<TProps>) => {
    return <div className="fixed right-6 bottom-10 z-50 text-nord0 dark:text-nord4">
        {props.notifications.map((notification, index) =>
            <Notification
                key={notification.uuid}
                notification={notification} />)}
        {/* <Notification
            danger>
            There have been errors creating your session
        </Notification>
        <Notification
            success>
            Your session has been created
        </Notification> */}
    </div>;
}