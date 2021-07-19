import { store } from '@polo/common/state/models';
import { IModal } from '@polo/common/state/models/modal-model';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useLayoutEffect } from 'react';
import { createPortal } from 'react-dom';
import { useModal } from './modal-hooks';
import './modal.scss';

export const ModalContainer = observer((props: React.PropsWithChildren<{}>) => {

    const { hide } = useModal();

    useLayoutEffect(() => {
        const listener = (e: KeyboardEvent) => {
            if (e.key === 'Escape') hide();
        };
        document.addEventListener('keydown', listener);
        return () => document.removeEventListener('keydown', listener);
    }, []);

    const disposeModal = () =>
        store.app.modal.setVisible(false);

    return <div className="modal" onClick={() => disposeModal()}>
        <div className="__content" onClick={e => e.stopPropagation()}>
            {props.children}
        </div>
    </div>;
})

export const ModalPortal = observer((props: React.PropsWithChildren<{ modal: IModal, name: string }>) => {

    return (props.modal.visible &&
        props.modal.name === props.name &&
        createPortal(<ModalContainer>
            {props.children}
        </ModalContainer>, document.getElementById('modal'))) || null;
});

export const Modal = (props: React.PropsWithChildren<{ name: string }>) => {

    const { hide } = useModal();

    useEffect(() => hide, []);

    return <ModalPortal modal={store.app.modal} name={props.name}>{props.children}</ModalPortal>
}