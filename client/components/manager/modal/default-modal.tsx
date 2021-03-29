import React from 'react';
import './default-modal.scss';
import { useModal } from './modal-hooks';
import { Modal } from './modal-portal';

export const DefaultModal = (props: React.PropsWithChildren<{ name: string; }>) => {

    const { hide } = useModal();

    return <Modal name={props.name}>
        <div className="default-modal">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 20 20"
                fill="currentColor"
                className="w-6 h-6 lg:w-5 lg:h-5 absolute right-2 top-2 cursor-pointer z-20"
                onClick={hide}>
                <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
            </svg>
            {props.children}
        </div>
    </Modal>
}