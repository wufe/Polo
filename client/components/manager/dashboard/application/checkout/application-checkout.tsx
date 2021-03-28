import React from 'react';
import { DefaultModal } from '@/components/manager/modal/default-modal';
import { useModal } from '@/components/manager/modal/modal-hooks';
import dayjs from 'dayjs';
import './application-checkout.scss';
import { CommitMessage } from '@/components/manager/shared/commit-message';
import { CheckoutBuildConfirmationModal } from './modal/checkout-build-confirmation-modal';

type TProps = {
    name                       : string;
    message                    : string;
    author                     : string;
    authorEmail                : string;
    date                       : string;
    onSessionCreationSubmission: (checkout: string) => void;
}
export const ApplicationCheckout = (props: TProps) => {

    const { show, hide } = useModal();
    const checkoutOptionsModalName = `checkout-${props.name}`;
    const checkoutBuildConfirmationModalName = `${checkoutOptionsModalName}-build-confirmation`;

    return <div
        className="application-checkout">
        <div className="__content" onClick={() => show(checkoutBuildConfirmationModalName)}>
            <div className="__title-container">
                <span
                    className="__title">
                    {props.name}
                </span>
                <div className="__subtitle-container">
                    <span className="__subtitle-item">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            className="w-3 h-3 mr-1">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                        </svg>
                        <span className="whitespace-nowrap">{props.author}</span>
                    </span>
                    <span className="__subtitle-item">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            className="w-3 h-3 mr-1">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <span className="whitespace-nowrap">{dayjs(props.date).format('DD MMM HH:mm')}</span>
                    </span>
                </div>
            </div>
        </div>
        <span className="text-center whitespace-nowrap flex flex-nowrap items-start">
            <span className="__button --success --hide-on-mobile" onClick={() => props.onSessionCreationSubmission(props.name)}>
                <span>Create</span>
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                </svg>
            </span>
            <span className="__button --ghost inline-flex" onClick={() => show(checkoutOptionsModalName)}>
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M6 10a2 2 0 11-4 0 2 2 0 014 0zM12 10a2 2 0 11-4 0 2 2 0 014 0zM16 12a2 2 0 100-4 2 2 0 000 4z" />
                </svg>
            </span>
        </span>
        <DefaultModal name={checkoutOptionsModalName} />
        <CheckoutBuildConfirmationModal
            name={checkoutBuildConfirmationModalName}
            commitAuthor={props.author}
            commitAuthorEmail={props.authorEmail}
            commitDate={props.date}
            commitMessage={props.message}
            onSessionCreationSubmission={props.onSessionCreationSubmission} />
    </div>
}