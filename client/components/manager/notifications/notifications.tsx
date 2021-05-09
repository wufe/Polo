import React from 'react';
import { store } from '@/state/models';
import { NotificationsPortal } from './notifications-portal';

export const Notifications = () =>
    <NotificationsPortal app={store.app} />;