import { DefaultModal } from '@/components/modal/default-modal';
import { useModal } from '@/components/modal/modal-hooks';
import { CommitMessage } from '@/components/shared/commit-message';
import { Button } from '@polo/common/components/elements/button/button';
import { CubeIcon } from '@polo/common/components/elements/icons/cube/cube-icon';
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
                <Button
                    success
                    outlined
                    icon={<CubeIcon />}
                    label="Create session"
                    onClick={() => {
                        hide();
                        props.onSessionCreationSubmission(props.checkoutName);
                    }} />
            </div>
        </div>
    </DefaultModal>;
}