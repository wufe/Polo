import { IApplication } from '@polo/common/state/models';
import { observer } from 'mobx-react-lite';
import React, { useEffect } from 'react';
import { useApplicationNotifications } from './application-notification-hook';

type TProps = {
    application: IApplication;
}
export const ApplicationNotifications = observer(({ application }: TProps) => {

    useApplicationNotifications(application);

    return <></>;
})