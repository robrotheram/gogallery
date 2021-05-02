import React from 'react';
import { PlusOutlined } from '@ant-design/icons';
import { Tag, Input, Tooltip } from 'antd';

export default class EditableTagGroup extends React.Component {
  constructor(props){
    super(props)
      this.state = {
        tags: [],
        removedTags: [],
        inputVisible: false,
        inputValue: '',
      };
  }

  static getDerivedStateFromProps(nextProps, prevState){
    if(nextProps.value === undefined || nextProps.value === null){return null}
    let intersection = nextProps.value.filter(x => !prevState.removedTags.includes(x));
    console.log("NEW PROPS", nextProps.value, prevState.removedTags, intersection)
    
      return { tags: intersection }
   
  }

  handleClose = removedTag => {
    const tags = this.state.tags.filter(tag => tag !== removedTag);
    console.log("CLOS TAG", tags);
    this.setState({ tags,
      removedTags: this.state.removedTags.concat(removedTag)
    });
    this.props.onChange(tags)
  };

  showInput = () => {
    this.setState({ inputVisible: true }, () => this.input.focus());
  };

  handleInputChange = e => {
    this.setState({ inputValue: e.target.value });
  
  };

  handleInputConfirm = () => {
    const { inputValue } = this.state;
    let { tags } = this.state;
    if (inputValue && tags.indexOf(inputValue) === -1) {
      tags = [...tags, inputValue];
    }
    this.setState({
      tags,
      inputVisible: false,
      inputValue: '',
    });
    console.log("Received values of form",tags, this.state)
    this.props.onChange(tags)
  };

  saveInputRef = input => (this.input = input);

  render() {
    const { tags, inputVisible, inputValue } = this.state;
    return (
      <div>
        {tags.map((tag, index) => {
          const isLongTag = tag.length > 20;
          const tagElem = (
            <Tag key={tag} closable={true} onClose={() => this.handleClose(tag)}>
              {isLongTag ? `${tag.slice(0, 20)}...` : tag}
            </Tag>
          );
          return isLongTag ? (
            <Tooltip title={tag} key={tag}>
              {tagElem}
            </Tooltip>
          ) : (
            tagElem
          );
        })}
        {inputVisible && (
          <Input
            ref={this.saveInputRef}
            type="text"
            size="small"
            style={{ width: 78 }}
            value={inputValue}
            onChange={this.handleInputChange}
            onBlur={this.handleInputConfirm}
            onPressEnter={this.handleInputConfirm}
          />
        )}
        {!inputVisible && (
          <Tag onClick={this.showInput} style={{ background: '#141414', borderStyle: 'dashed' }}>
            <PlusOutlined /> New Tag
          </Tag>
        )}
      </div>
    );
  }
}