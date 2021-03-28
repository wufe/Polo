import React from 'react';
import { DefaultModal } from '@/components/manager/modal/default-modal';
import { ISession } from '@/state/models/session-model';
import './application-session-deletion-modal.scss';

type TProps = {
    name                         : string;
    session                      : ISession;
    onApplicationDeletionSelected: () => void;
}
export const ApplicationSessionDeletionModal = (props: TProps) => {
    return <DefaultModal name={props.name}>
        <div className="application-session-deletion-modal">
            <div className="__header">
                <div className="text-base lg:text-lg font-bold whitespace-nowrap overflow-hidden overflow-ellipsis">{props.session.checkout}</div>
                <div className="text-xs text-gray-500 opacity-80">{props.session.uuid}</div>
            </div>
            <div className="__description">
                You are going to delete the session. Are you sure?
            </div>
            <div className="__actions-container mt-5 flex justify-center">
                <span className="__button --danger --outlined" onClick={props.onApplicationDeletionSelected}>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                    <span>Delete</span>
                </span>
            </div>
        </div>
    </DefaultModal>
}