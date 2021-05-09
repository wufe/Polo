import { ISession, TNotificationProps } from "../models";
import { NotificationType } from "../models/notification-model";

export const buildFailedNotification = (session: ISession, onClick: TNotificationProps['onClick']): TNotificationProps => ({
    text: 'Build failed',
    type: NotificationType.ERROR,
    onClick
});