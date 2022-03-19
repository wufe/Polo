import {memo} from 'preact/compat';
import { HelperStatusContext } from '../contexts';

const SessionAgeInterval = memo((props: { age: number }) => {
    return <>Expires in <b>{props.age}s</b></>;
}, (prev, next) => prev.age === next.age);

export const SessionAge = () => {
    return <HelperStatusContext.Consumer>
        {({ age }) => <SessionAgeInterval age={age} />}
    </HelperStatusContext.Consumer>
}