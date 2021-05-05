import { DefaultModal } from '@/components/manager/modal/default-modal';
import { useModal } from '@/components/manager/modal/modal-hooks';
import { CommitMessage } from '@/components/manager/shared/commit-message';
import React from 'react';
import './checkout-build-confirmation-modal.scss';

type TProps = {
    name                       : string;
    checkoutName               : string;
    commitAuthor               : string;
    commitAuthorEmail          : string;
    commitDate                 : string;
    commitMessage              : string;
    onSessionCreationSubmission: (checkout: string) => void;
}
export const CheckoutBuildConfirmationModal = (props: TProps) => {

    const { hide } = useModal();

    return <DefaultModal name={props.name}>
        <div className="checkout-build-confirmation-modal">
            <div className="__header mb-6">
                <div className="text-base lg:text-lg font-bold whitespace-nowrap overflow-hidden overflow-ellipsis">{props.checkoutName}</div>
            </div>
            <CommitMessage
                commitAuthorEmail={props.commitAuthorEmail}
                commitAuthorName={props.commitAuthor}
                commitDate={props.commitDate}
                commitMessage={props.commitMessage} />
            <div className="__actions-container">
                <span className="__button --success --outlined" onClick={() => {
                    hide();
                    props.onSessionCreationSubmission(props.checkoutName);
                }}>
                    <span>Create session</span>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                    </svg>
                </span>
            </div>
        </div>
    </DefaultModal>;
}