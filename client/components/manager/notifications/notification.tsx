import React, { MouseEventHandler } from 'react';
import './notification.scss';
import classnames from 'classnames';
import { XIcon } from '@heroicons/react/outline'
import { INotification, NotificationType } from '@/state/models/notification-model';

type TProps = {
    notification: INotification;
};
export const Notification = ({
    notification: {
        expiration,
        text,
        title,
        type,
        triggerOnClick,
        remove,
    }
}: React.PropsWithChildren<TProps>) => {

    if (!title) {
        switch (type) {
            case NotificationType.ERROR:
                title = 'Error';
                break;
            case NotificationType.INFO:
                title = 'Info';
                break;
            case NotificationType.SUCCESS:
                title = 'Success';
                break;
        }
    }

    const onCloseClick: MouseEventHandler<HTMLDivElement> = event => {
        event.stopPropagation();
        remove();
    }

    return <div className={classnames('notification', {
        '--success': type === NotificationType.SUCCESS,
        '--danger' : type === NotificationType.ERROR,
        '--info'   : type === NotificationType.INFO,
    })} onClick={triggerOnClick}>
        <div className="__title">{title}</div>
        <div style={{ display: 'flex', flexDirection: 'column' }}>
            {text.split('\n').map((line, index) => <span key={index}>
                {line}
                <br />
            </span>)}
        </div>
        <div className="__loading-bar" style={{
            animationDuration: `${expiration}s`
        }}></div>
        <div className="__icon-wrapper" onClick={onCloseClick}>
            <XIcon className="w-4 h-4" />
        </div>
    </div>
}