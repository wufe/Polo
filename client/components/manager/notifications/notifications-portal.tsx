import React from 'react';
import { createPortal } from 'react-dom';
import { observer } from 'mobx-react-lite';
import { NotificationsContainer } from './notifications-container';
import { IApp } from '@/state/models/app-model';
import { values } from 'mobx';
import { INotification } from '@/state/models/notification-model';

type TProps = {
    app: IApp;
};
export const NotificationsPortal = observer((props: React.PropsWithChildren<TProps>) => {
    return createPortal(<NotificationsContainer
        notifications={values(props.app.notifications) as any as INotification[]} />, document.getElementById('notifications'));
});