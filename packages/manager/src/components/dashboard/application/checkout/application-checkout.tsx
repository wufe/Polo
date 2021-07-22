import React from 'react';
import { DefaultModal } from '@/components/modal/default-modal';
import { useModal } from '@/components/modal/modal-hooks';
import dayjs from 'dayjs';
import './application-checkout.scss';
import { CommitMessage } from '@/components/shared/commit-message';
import { CheckoutBuildConfirmationModal } from './modal/checkout-build-confirmation-modal';
import { ApplicationCheckoutModal } from './modal/application-checkout-modal';
import { CommitModal } from '@/components/shared/commit-modal';
import { Button } from '@polo/common/components/elements/button/button';
import { CubeIcon } from '@polo/common/components/elements/icons/cube/cube-icon';
import { HorizontalDotsIcon } from '@polo/common/components/elements/icons/horizontal-dots/horizontal-dots-icon';

type TProps = {
    type                       : 'branch' | 'tag';
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
    const commitMessageModalName = `${checkoutOptionsModalName}-commit`;
    const checkoutBuildConfirmationModalName = `${checkoutOptionsModalName}-build-confirmation`;

    return <div
        className="application-checkout">
        <a className="__content" onClick={() => show(checkoutBuildConfirmationModalName)}>
            <div className="w-6 flex justify-center items-center mr-1">
                {props.type === 'branch' && <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="w-4 h-4 mt-1"
                    viewBox="0 0 512 512"
                    fill="currentColor">
                    <path d="M416 160a64 64 0 10-96.27 55.24c-2.29 29.08-20.08 37-75 48.42-17.76 3.68-35.93 7.45-52.71 13.93v-126.2a64 64 0 10-64 0v209.22a64 64 0 1064.42.24c2.39-18 16-24.33 65.26-34.52 27.43-5.67 55.78-11.54 79.78-26.95 29-18.58 44.53-46.78 46.36-83.89A64 64 0 00416 160zM160 64a32 32 0 11-32 32 32 32 0 0132-32zm0 384a32 32 0 1132-32 32 32 0 01-32 32zm192-256a32 32 0 1132-32 32 32 0 01-32 32z"></path>
                </svg>}
                {props.type === 'tag' && <svg
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    className="w-4 h-4 mt-0.5"
                    viewBox="0 0 24 24"
                    stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                </svg>}
            </div>
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
        </a>
        <span className="text-center whitespace-nowrap flex flex-nowrap items-start">
            <Button
                ghost
                onClick={() => props.onSessionCreationSubmission(props.name)}
                label="Create"
                icon={<CubeIcon />} />
            <Button
                ghost
                onClick={() => show(checkoutOptionsModalName)}
                icon={<HorizontalDotsIcon />} />
        </span>
        <ApplicationCheckoutModal
            name={checkoutOptionsModalName}
            checkoutName={props.name}
            onSessionCreationSubmission={() => {
                hide();
                props.onSessionCreationSubmission(props.name);
            }}
            onCommitMessageSelection={() => show(commitMessageModalName)} />
        <CommitModal
            name={commitMessageModalName}
            title={props.name}
            commitAuthorEmail={props.authorEmail}
            commitAuthorName={props.author}
            commitDate={props.date}
            commitMessage={props.message} />
        <CheckoutBuildConfirmationModal
            name={checkoutBuildConfirmationModalName}
            checkoutName={props.name}
            commitAuthor={props.author}
            commitAuthorEmail={props.authorEmail}
            commitDate={props.date}
            commitMessage={props.message}
            onSessionCreationSubmission={props.onSessionCreationSubmission} />
    </div>
}