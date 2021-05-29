import React from 'react';
import { DefaultModal } from '@/components/manager/modal/default-modal';
import { ISession } from '@/state/models/session-model';
import './application-session-deletion-modal.scss';
import { Button } from '@/components/shared/elements/button/button';
import { TrashIcon } from '@/components/shared/elements/icons/trash/trash-icon';

type TProps = {
    name                         : string;
    session                      : ISession;
    onApplicationDeletionSelected: () => void;
}
export const ApplicationSessionDeletionModal = (props: TProps) => {
    return <DefaultModal name={props.name}>
        <div className="application-session-deletion-modal">
            <div className="__header">
                <div className="text-base lg:text-lg font-bold whitespace-nowrap overflow-hidden overflow-ellipsis">{props.session.displayName}</div>
                <div className="text-xs text-gray-500 dark:text-gray-400 opacity-80">{props.session.uuid}</div>
            </div>
            <div className="__description">
                You are going to delete the session. Are you sure?
            </div>
            <div className="__actions-container mt-5 flex justify-center">
                <Button
                    danger
                    outlined
                    label="Delete"
                    onClick={props.onApplicationDeletionSelected}
                    icon={<TrashIcon />} />
            </div>
        </div>
    </DefaultModal>
}