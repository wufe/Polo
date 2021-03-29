import { DefaultModal } from '@/components/manager/modal/default-modal';
import React from 'react';
import './application-checkout-modal.scss';

type TProps = {
    name                       : string;
    checkoutName               : string;
    onSessionCreationSubmission: () => void;
    onCommitMessageSelection      : () => void;
}
export const ApplicationCheckoutModal = (props: TProps) => {
    return <DefaultModal name={props.name}>
        <div className="application-checkout-modal">
            <div className="__header">
                <div className="text-base lg:text-lg font-bold whitespace-nowrap overflow-hidden overflow-ellipsis">{props.checkoutName}</div>
            </div>
            <div className="__list">
                <div className="__item --show-on-mobile" onClick={props.onSessionCreationSubmission}>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
                    </svg>
                    <span>Create session</span>
                </div>
                <div className="flex justify-center my-2 sm:hidden">
                    <div className="border-t border-gray-500 w-full opacity-40" style={{ height: 1 }}></div>
                </div>
                <div className="__item" onClick={props.onCommitMessageSelection}>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
                    </svg>
                    <span>View commit message</span>
                </div>
            </div>
        </div>
    </DefaultModal>;
}