import React from 'react';
import { store } from '@polo/common/state/models';
import { NotificationsPortal } from './notifications-portal';

export const Notifications = () =>
    <NotificationsPortal app={store.app} />;